// MIT License
//
// (C) Copyright [2018-2023,2025] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-certs/pkg/hms_certs"
	compcreds "github.com/Cray-HPE/hms-compcredentials"
	sstorage "github.com/Cray-HPE/hms-securestorage"
	jwtauth "github.com/OpenCHAMI/jwtauth/v5"
	"github.com/OpenCHAMI/smd/v2/internal/hbtdapi"
	"github.com/OpenCHAMI/smd/v2/internal/hmsds"
	"github.com/OpenCHAMI/smd/v2/internal/pgmigrate"
	"github.com/OpenCHAMI/smd/v2/internal/slsapi"
	rf "github.com/OpenCHAMI/smd/v2/pkg/redfish"
	"github.com/OpenCHAMI/smd/v2/pkg/sm"
	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

type SmdFlavor string

const (
	OpenCHAMI        = "OpenCHAMI"
	CSM              = "CSM"
	UnknownSmdFlavor = "UnknownSmdFlavor"
)

const (
	dbTypeMySQL    = "mysql" // No longer supported
	dbTypePostgres = "postgres"
)

const (
	SCNMAP_ENABLED  = 0
	SCNMAP_ROLE     = 1
	SCNMAP_SUBROLE  = 2
	SCNMAP_SWSTATUS = 3
	SCNMAP_STATE    = 4
	SCNMAP_MAX      = 5
)

type SCNUrl struct {
	url      string
	refCount int
}

type SCNSubMap [SCNMAP_MAX]map[string][]SCNUrl

type Job struct {
	job        *sm.Job
	cancelChan chan bool
}

const httpListenDefault = ":27779"

type SmD struct {
	db    hmsds.HMSDB
	dbDSN string

	dbType    string
	dbName    string
	dbUser    string
	dbPass    string
	dbHost    string
	dbPortStr string
	dbPort    int
	dbOpts    string

	logDir          string
	tlsCert         string
	tlsKey          string
	proxyURL        string
	httpListen      string
	msgbusListen    string
	logLevelIn      int
	msgbusConfig    MsgBusConfigWrapper
	msgbusHandle    MsgbusHandleWrapper
	hwInvHistAgeMax int
	smapCompEP      *SyncMap
	genTestPayloads string
	enableDiscovery bool
	openchami       bool
	zerolog         bool

	// v2 APIs
	apiRootV2           string
	serviceBaseV2       string
	valuesBaseV2        string
	stateBaseV2         string
	componentsBaseV2    string
	redfishEPBaseV2     string
	compEPBaseV2        string
	serviceEPBaseV2     string
	compEthIntBaseV2    string
	hsnIntBaseV2        string
	hwinvByLocBaseV2    string
	hwinvByFRUBaseV2    string
	invDiscoverBaseV2   string
	invDiscStatusBaseV2 string
	nodeMapBaseV2       string
	subscriptionBaseV2  string
	groupsBaseV2        string
	partitionsBaseV2    string
	membershipsBaseV2   string
	compLockBaseV2      string
	sysInfoBaseV2       string
	powerMapBaseV2      string

	wp            *base.WorkerPool
	wpRFEvent     *base.WorkerPool
	scnSubs       sm.SCNSubscriptionArray
	scnSubMap     SCNSubMap
	scnSubLock    sync.Mutex
	lg            *log.Logger // Log file
	lgLvl         LogLevel
	slsUrl        string
	sls           *slsapi.SLS
	hbtdUrl       string
	hbtd          *hbtdapi.HBTD
	hmsConfigPath string

	// TODO: Remove anything conditional on writeVault when HSM no longer is
	//       the one writing credentials to Vault.
	writeVault bool
	readVault  bool
	ss         sstorage.SecureStorage
	ccs        *compcreds.CompCredStore

	// Job Sync
	jobLock     sync.Mutex
	jobList     map[string]*Job // List of Jobs by jobId running on this HSM instance.
	srfpJobList map[string]*Job // List of srfp Jobs by xname running on this HSM instance.

	//Discovery Sync
	discMap     map[string]int
	discMapLock sync.Mutex

	//router
	router    *chi.Mux
	tokenAuth *jwtauth.JWTAuth
	jwksURL   string

	httpClient *retryablehttp.Client
}

type LogLevel int

const (
	LOG_DEFAULT LogLevel = 0
	LOG_NOTICE  LogLevel = 1
	LOG_INFO    LogLevel = 2
	LOG_DEBUG   LogLevel = 3
	LOG_LVL_MAX LogLevel = 4
)

var serviceName string

func getSmdFlavor() (smdFlavor SmdFlavor, moduleName string) {
	info, _ := debug.ReadBuildInfo()
	moduleName = info.Path
	m := strings.ToLower(moduleName)
	if strings.HasPrefix(m, "github.com/openchami") {
		return OpenCHAMI, moduleName
	} else if strings.HasPrefix(m, "github.com/cray-hpe") {
		return CSM, moduleName
	}
	return UnknownSmdFlavor, moduleName
}

func (s *SmD) Log(lvl LogLevel, format string, a ...interface{}) {
	if int(lvl) <= int(s.lgLvl) {
		// depth=2, get line num of caller, not us
		s.lg.Output(2, fmt.Sprintf(format, a...))
	}
}

func (s *SmD) LogAlwaysStr(format string) {
	s.lg.Output(2, format)
}

func (s *SmD) LogAlways(format string, a ...interface{}) {
	// Use caller's line number (depth=2)
	s.lg.Output(2, fmt.Sprintf(format, a...))
}

func (s *SmD) SetLogLevel(lvl LogLevel) error {
	if lvl >= LOG_DEFAULT && lvl < LOG_LVL_MAX {
		s.lgLvl = lvl
		return nil
	} else {
		err := errors.New("warning: verbose level unchanged")
		s.lg.Printf("%s", err)
		return err
	}
}

// Add a SCN URL to the specified list of unique URLs. If a duplicate exists,
// the refCount is increased.
func addSCNUrl(urlList []SCNUrl, urlIn string) []SCNUrl {
	found := false
	for i, url := range urlList {
		if urlIn == url.url {
			found = true
			urlList[i].refCount++
			break
		}
	}
	if !found {
		url := SCNUrl{url: urlIn, refCount: 1}
		urlList = append(urlList, url)
	}
	return urlList
}

// Remove a SCN URL from the specified list of unique URLs. URLs are not
// removed from the list until the refCount is < 1.
func removeSCNUrl(urlList []SCNUrl, urlIn string) []SCNUrl {
	for i, url := range urlList {
		if url.url == urlIn {
			if url.refCount <= 1 {
				urlList = append(urlList[:i], urlList[i+1:]...)
			} else {
				urlList[i].refCount--
			}
		}
	}
	return urlList
}

// Add a SCN subscription to the specified SCN subscription map
func addSCNMapSubscription(subMap *SCNSubMap, sub *sm.SCNSubscription) {
	if sub.Enabled != nil && *sub.Enabled {
		if subMap[SCNMAP_ENABLED] == nil {
			subMap[SCNMAP_ENABLED] = make(map[string][]SCNUrl, 0)
		}
		if _, ok := subMap[SCNMAP_ENABLED]["enabled"]; !ok {
			subMap[SCNMAP_ENABLED]["enabled"] = make([]SCNUrl, 0, 1)
		}
		subMap[SCNMAP_ENABLED]["enabled"] = addSCNUrl(subMap[SCNMAP_ENABLED]["enabled"], sub.Url)
	}
	for _, rl := range sub.Roles {
		role := strings.ToLower(rl)
		if subMap[SCNMAP_ROLE] == nil {
			subMap[SCNMAP_ROLE] = make(map[string][]SCNUrl, 0)
		}
		if _, ok := subMap[SCNMAP_ROLE][role]; !ok {
			subMap[SCNMAP_ROLE][role] = make([]SCNUrl, 0, 1)
		}
		subMap[SCNMAP_ROLE][role] = addSCNUrl(subMap[SCNMAP_ROLE][role], sub.Url)
	}
	for _, srl := range sub.SubRoles {
		subRole := strings.ToLower(srl)
		if subMap[SCNMAP_SUBROLE] == nil {
			subMap[SCNMAP_SUBROLE] = make(map[string][]SCNUrl, 0)
		}
		if _, ok := subMap[SCNMAP_SUBROLE][subRole]; !ok {
			subMap[SCNMAP_SUBROLE][subRole] = make([]SCNUrl, 0, 1)
		}
		subMap[SCNMAP_SUBROLE][subRole] = addSCNUrl(subMap[SCNMAP_SUBROLE][subRole], sub.Url)
	}
	for _, swst := range sub.SoftwareStatus {
		swStatus := strings.ToLower(swst)
		if subMap[SCNMAP_SWSTATUS] == nil {
			subMap[SCNMAP_SWSTATUS] = make(map[string][]SCNUrl, 0)
		}
		if _, ok := subMap[SCNMAP_SWSTATUS][swStatus]; !ok {
			subMap[SCNMAP_SWSTATUS][swStatus] = make([]SCNUrl, 0, 1)
		}
		subMap[SCNMAP_SWSTATUS][swStatus] = addSCNUrl(subMap[SCNMAP_SWSTATUS][swStatus], sub.Url)
	}
	for _, st := range sub.States {
		state := strings.ToLower(st)
		if subMap[SCNMAP_STATE] == nil {
			subMap[SCNMAP_STATE] = make(map[string][]SCNUrl, 0)
		}
		if _, ok := subMap[SCNMAP_STATE][state]; !ok {
			subMap[SCNMAP_STATE][state] = make([]SCNUrl, 0, 1)
		}
		subMap[SCNMAP_STATE][state] = addSCNUrl(subMap[SCNMAP_STATE][state], sub.Url)
	}
}

// Remove a SCN subscription from the specified SCN subscription map
func removeSCNMapSubscription(subMap *SCNSubMap, sub *sm.SCNSubscription) {
	if sub.Enabled != nil && *sub.Enabled {
		subMap[SCNMAP_ENABLED]["enabled"] = removeSCNUrl(subMap[SCNMAP_ENABLED]["enabled"], sub.Url)
	}
	for _, rl := range sub.Roles {
		role := strings.ToLower(rl)
		subMap[SCNMAP_ROLE][role] = removeSCNUrl(subMap[SCNMAP_ROLE][role], sub.Url)
	}
	for _, srl := range sub.SubRoles {
		subRole := strings.ToLower(srl)
		subMap[SCNMAP_SUBROLE][subRole] = removeSCNUrl(subMap[SCNMAP_SUBROLE][subRole], sub.Url)
	}
	for _, swst := range sub.SoftwareStatus {
		swStatus := strings.ToLower(swst)
		subMap[SCNMAP_SWSTATUS][swStatus] = removeSCNUrl(subMap[SCNMAP_SWSTATUS][swStatus], sub.Url)
	}
	for _, st := range sub.States {
		state := strings.ToLower(st)
		subMap[SCNMAP_STATE][state] = removeSCNUrl(subMap[SCNMAP_STATE][state], sub.Url)
	}
}

// Spin off a thread to periodically refresh the SCN subscription tables.
func (s *SmD) SCNSubscriptionRefresh() {
	go func() {
		for {
			s.scnSubLock.Lock()
			subs, err := s.db.GetSCNSubscriptionsAll()
			if err != nil {
				s.scnSubLock.Unlock()
				s.LogAlways("SCNSubscriptionRefresh(): Lookup failure: %s", err)
				time.Sleep(10 * time.Second)
			} else {
				//TODO: Make this only refresh if there was a change in the db not made by this instance of HSM.
				// Refresh the internal subscription list and map
				newSCNSubMap := SCNSubMap{}
				for _, sub := range subs.SubscriptionList {
					addSCNMapSubscription(&newSCNSubMap, &sub)
				}
				s.scnSubs = *subs
				s.scnSubMap = newSCNSubMap
				s.scnSubLock.Unlock()
				time.Sleep(30 * time.Second)
			}
		}
	}()
}

// Spin off a thread to periodically clean up expired component locks.
func (s *SmD) CompReservationCleanup() {
	go func() {
		for {
			xnames, err := s.db.DeleteCompReservationsExpired()
			if err != nil {
				s.LogAlways("CompReservationCleanup(): Lookup failure: %s", err)
				time.Sleep(10 * time.Second)
			} else {
				if len(xnames) > 0 {
					s.LogAlways("CompReservationCleanup(): Release %d expired component reservations for: %v", len(xnames), xnames)
				}
				time.Sleep(30 * time.Second)
			}
		}
	}()
}

// Jobs running locally in an intance of HSM can become orphaned if that
// instance of HSM dies. This spins off a goroutine to periodically check for
// orphaned jobs and picks them up.
func (s *SmD) JobSync() {
	go func() {
		for {
			failed := false
			numNewJobs := 0
			// Take on orphaned Jobs
			jobs, err := s.db.GetJobs(hmsds.JS_Expired)
			if err != nil {
				s.LogAlways("JobSync(): Lookup failure: %s", err)
				failed = true
			} else {
				for _, job := range jobs {
					if job.Status == sm.JobComplete {
						s.db.DeleteJob(job.Id)
					} else {
						// Delete and remake the job. This will make it so only
						// one HSM instance can pick up the job.
						didDelete, err := s.db.DeleteJob(job.Id)
						if err != nil || !didDelete {
							// Don't create the new job if we didn't delete a job
							continue
						}
						switch job.Type {
						case sm.JobTypeSRFP:
							data, ok := job.Data.(*sm.SrfpJobData)
							if ok {
								// Start orphaned SRFP jobs without the initial delay
								s.doStateRFPoll(data.CompId, 0)
							}
						}
						numNewJobs++
					}
				}
			}
			if failed {
				// Don't wait the full 15 seconds to retry reading from the database
				time.Sleep(10 * time.Second)
			} else {
				if numNewJobs > 0 {
					s.LogAlways("JobSync(): Picked up %d orphaned jobs", numNewJobs)
				}
				time.Sleep(20 * time.Second)
			}
		}
	}()
}

// Discovery jobs running locally in an intance of HSM can become orphaned if that
// instance of HSM dies. This spins off a goroutine to periodically check for
// orphaned discovery jobs and picks them up.
// TODO: Close potential race condition between GetRFEndpointsFilter() and when
//
//	discoverFromEndpoint() updates the DiscInfo.LastAttempt where another HSM
//	instance could run GetRFEndpointsFilter() to get the same list of orphaned
//	EPs. For now, this should be rare.
func (s *SmD) DiscoverySync() {
	go func() {
		for {
			failed := false
			numNewJobs := 0
			// Gather a list of all in-progress discovery jobs
			eps, err := s.db.GetRFEndpointsFilter(&hmsds.RedfishEPFilter{LastStatus: []string{rf.DiscoveryStarted}})
			if err != nil {
				s.LogAlways("DiscoverySync(): Lookup failure: %s", err)
				failed = true
			} else {
				for _, ep := range eps {
					lastAttempt, _ := time.Parse("2006-01-02T15:04:05.000000Z07:00", ep.DiscInfo.LastAttempt)
					// Consider discovery jobs that have not updated
					// in 30 minutes to have been orphaned.
					if time.Since(lastAttempt) >= (time.Minute * 30) {
						// Take on orphaned discovery job
						go s.discoverFromEndpoint(ep, 0, true)
						numNewJobs++
					}
				}
			}
			if failed {
				// Don't wait the full time to retry reading from the database
				time.Sleep(30 * time.Second)
			} else {
				if numNewJobs > 0 {
					s.LogAlways("DiscoverySync(): Picked up %d orphaned discovery jobs", numNewJobs)
				}
				// Check every 10 minutes
				time.Sleep(10 * time.Minute)
			}
		}
	}()
}

// Add an xname to the discovery list. Meaning HSM has started
// discovery on this component and a discovery job is ongoing
// and owned by this HSM instance.
func (s *SmD) discoveryMapAdd(id string) {
	s.discMapLock.Lock()
	ref, ok := s.discMap[id]
	if ok {
		s.discMap[id] = ref + 1
	} else {
		s.discMap[id] = 1
	}
	s.discMapLock.Unlock()
}

// Remove an xname from the discovery list. Meaning HSM has finished
// discovery on this component.
func (s *SmD) discoveryMapRemove(id string) {
	s.discMapLock.Lock()
	ref, ok := s.discMap[id]
	if ok {
		if ref <= 1 {
			delete(s.discMap, id)
		} else {
			s.discMap[id] = ref - 1
		}
	}
	s.discMapLock.Unlock()
}

// Periodically updates the timestamp for ongoing RedfishEndpoint discoveries
// owned by this instance of HSM so that a discovery that is taking longer than
// usual doesn't get seen as an orphan by another HSM instance.
func (s *SmD) DiscoveryUpdater() {
	go func() {
		for {
			s.discMapLock.Lock()
			if len(s.discMap) > 0 {
				discIDs := make([]string, 0, 1)
				for id := range s.discMap {
					discIDs = append(discIDs, id)
				}
				// Cause the discovery LastStatus to get updated for all of these IDs to
				// let the other HSM instances know that we're still working on them.
				_, err := s.db.UpdateRFEndpointForDiscover(discIDs, true)
				if err != nil {
					s.discMapLock.Unlock()
					// Don't wait the full time to retry updating the database
					time.Sleep(30 * time.Second)
					continue
				}
			}
			s.discMapLock.Unlock()
			// Update every 20 minutes. This is a bit arbitrary. 20mins is long
			// enough that, on average, the discovery should have finished thus
			// no need to update the timestamp but shorter than the 30 min
			// timeout for when other HSM instances mark a discovery job as
			// orphaned.
			time.Sleep(20 * time.Minute)
		}
	}()
}

func (s *SmD) GetHTTPClient() *retryablehttp.Client {
	if s.httpClient == nil {
		s.httpClient = retryablehttp.NewClient()
		s.httpClient.HTTPClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		s.httpClient.RetryMax = 5
		s.httpClient.HTTPClient.Timeout = time.Second * 40
		//turn off the http client loggin!
		tmpLogger := logrus.New()
		tmpLogger.SetLevel(logrus.PanicLevel)
		s.httpClient.Logger = tmpLogger
	}
	return s.httpClient
}

var applyMigrations bool

// Parse command line options.
func (s *SmD) parseCmdLine() {
	enableDiscoveryDefault := ENABLE_DISCOVERY_DEFAULT
	envvar := "ENABLE_DISCOVERY"
	if val := os.Getenv(envvar); val != "" {
		b, err := strconv.ParseBool(val)
		if err != nil {
			fmt.Printf("Warning: Bad env %s - '%s'\n", envvar, val)
		} else {
			// This is the default value for s.enableDiscovery if the cli option --enable-discovery was not used
			enableDiscoveryDefault = b
		}
	}

	flag.StringVar(&s.msgbusListen, "msg-host", "",
		"Host:Port:Topic for message bus. Not used if unset")
	flag.StringVar(&s.slsUrl, "sls-url", "",
		"Host:Port/base_path for communicating with SLS. Not used if unset")
	flag.StringVar(&s.hbtdUrl, "hbtd-url", "",
		"Host:Port/base_path for communicating with HBTD. Not used if unset")
	flag.StringVar(&s.proxyURL, "proxy", "",
		"Proxy to use (e.g. socks5://127.0.0.1:9999), none if unset (default)")
	flag.StringVar(&s.dbDSN, "db-dsn", "", "DSN to connect to database")
	flag.StringVar(&s.httpListen, "http-listen", httpListenDefault,
		"HTTP server IP + port binding")
	flag.StringVar(&s.tlsCert, "tls-cert", "/etc/cert.pem",
		"TLS cert file")
	flag.StringVar(&s.tlsKey, "tls-key", "/etc/key.pem",
		"TLS key file")
	flag.IntVar(&s.logLevelIn, "log", int(LOG_DEFAULT),
		"Log level: 0 to 4")
	flag.StringVar(&s.dbType, "dbtype", "",
		"Database type: 'mysql' (default) or 'postgres'")
	flag.StringVar(&s.dbName, "dbname", "", "Database name (default 'hmsds'")
	flag.StringVar(&s.dbUser, "dbuser", "", "Database user name")
	flag.StringVar(&s.dbHost, "dbhost", "", "Database hostname")
	flag.StringVar(&s.dbPortStr, "dbport", "", "Database port")
	flag.StringVar(&s.dbOpts, "dbopts", "", "Database options string")
	flag.StringVar(&s.jwksURL, "jwks-url", "", "Set the JWKS URL to fetch public key for validation")
	flag.BoolVar(&applyMigrations, "migrate", false, "Apply all database migrations before starting")
	flag.BoolVar(&s.enableDiscovery, "enable-discovery", enableDiscoveryDefault, "Enable discovery-related subroutines")
	flag.BoolVar(&s.openchami, "openchami", OPENCHAMI_DEFAULT, "Enabled OpenCHAMI features")
	flag.BoolVar(&s.zerolog, "zerolog", ZEROLOG_DEFAULT, "Enabled zerolog")
	help := flag.Bool("h", false, "Print help and exit")

	flag.Parse()

	// Help message.
	if help != nil && *help {
		flag.Usage()
		os.Exit(0)
	}
	envvar = "RF_MSG_HOST"
	if s.msgbusListen == "" {
		if val := os.Getenv(envvar); val != "" {
			s.msgbusListen = val
		}
	}
	envvar = "DBDSN"
	if s.dbDSN == "" {
		if val := os.Getenv(envvar); val != "" {
			s.dbDSN = val
		}
	}
	envvar = "SMD_PROXY"
	if s.proxyURL == "" {
		if val := os.Getenv(envvar); val != "" {
			s.proxyURL = val
		}
	}
	envvar = "SMD_DBTYPE"
	if s.dbType == "" {
		if val := os.Getenv(envvar); val != "" {
			s.dbType = val
		}
	}
	envvar = "SMD_DBNAME"
	if s.dbName == "" {
		if val := os.Getenv(envvar); val != "" {
			s.dbName = val
		}
	}
	envvar = "SMD_DBUSER"
	if s.dbUser == "" {
		if val := os.Getenv(envvar); val != "" {
			s.dbUser = val
		}
	}
	envvar = "SMD_DBHOST"
	if s.dbHost == "" {
		if val := os.Getenv(envvar); val != "" {
			s.dbHost = val
		}
	}
	envvar = "SMD_DBPORT"
	if s.dbPortStr == "" {
		if val := os.Getenv(envvar); val != "" {
			s.dbPortStr = val
		} else {
			s.dbPortStr = "5432"
		}
	}
	envvar = "SMD_JWKS_URL"
	if s.jwksURL == "" {
		if val := os.Getenv(envvar); val != "" {
			s.jwksURL = val
		}
	}

	port, err := strconv.ParseInt(s.dbPortStr, 10, 64)
	if err != nil {
		fmt.Printf("Bad dbport '%s': %s", s.dbPortStr, err)
		flag.Usage()
		os.Exit(1)
	}
	s.dbPort = int(port)

	envvar = "SMD_DBOPTS"
	if s.dbOpts == "" {
		if val := os.Getenv(envvar); val != "" {
			s.dbOpts = val
		}
	}
	// Env var only
	envvar = "GEN_TEST_PAYLOADS"
	if val := os.Getenv(envvar); val != "" {
		s.genTestPayloads = val
	}
	// Env var only
	envvar = "SMD_DBPASS"
	if val := os.Getenv(envvar); val != "" {
		s.dbPass = val
	}

	// Set dbName
	if s.dbName == "" {
		s.dbName = "hmsds"
	}

	// Set dbType
	if s.dbType == "" {
		s.dbType = dbTypePostgres
	} else if strings.ToLower(s.dbType) == dbTypePostgres {
		s.dbType = dbTypePostgres
	} else {
		fmt.Printf("Bad/missing dbtype\n")
		flag.Usage()
		os.Exit(1)
	}

	envvar = "SMD_WVAULT"
	if val := os.Getenv(envvar); val != "" {
		b, err := strconv.ParseBool(val)
		if err != nil {
			fmt.Printf("Warning: Bad env SMD_WVAULT - '%s'\n", val)
		} else {
			s.writeVault = b
		}
	}

	envvar = "SMD_RVAULT"
	if val := os.Getenv(envvar); val != "" {
		b, err := strconv.ParseBool(val)
		if err != nil {
			fmt.Printf("Warning: Bad env SMD_RVAULT - '%s'\n", val)
		} else {
			s.readVault = b
		}
	}

	s.hwInvHistAgeMax = 365
	envvar = "SMD_HWINVHIST_AGE_MAX_DAYS"
	if val := os.Getenv(envvar); val != "" {
		maxAge, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			fmt.Printf("Bad SMD_HWINVHIST_AGE_MAX_DAYS '%s': %s", val, err)
		} else if maxAge < 1 {
			fmt.Printf("Bad SMD_HWINVHIST_AGE_MAX_DAYS '%s': Must be 1+ days", val)
		} else {
			s.hwInvHistAgeMax = int(maxAge)
		}
	}

	s.hmsConfigPath = "/hms_config/hms_config.json"
	envvar = "HMS_CONFIG_PATH"
	if val := os.Getenv(envvar); val != "" {
		s.hmsConfigPath = val
	}

	if s.hbtdUrl == "" {
		envvar = "SMD_HBTD_HOST"
		if val := os.Getenv(envvar); val != "" {
			s.hbtdUrl = val
		}
	}

	// If DSN was not given, generate it from the individual DB options
	s.setDSN()
}

// Call DB-specific function to create DSN if an explicit one is not given.
func (s *SmD) setDSN() {
	if s.dbDSN != "" {
		return
	}
	if s.dbType == dbTypePostgres {
		s.dbDSN = hmsds.GenDsnHMSDB_PB(s.dbName, s.dbUser, s.dbPass,
			s.dbHost, s.dbOpts, s.dbPort)
	}
	if s.dbDSN == "" {
		fmt.Printf("Empty DSN created via flag or db options\n")
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	PrintVersionInfo()

	fmt.Printf("Build time defaults. MsgbusBuild: %t, RFEventMonitorBuild: %t, openChamiDefault: %t, zerologDefault: %t, enableDiscoveryDefault: %t\n",
		MSG_BUS_BUILD, RF_EVENT_MONITOR_BUILD, OPENCHAMI_DEFAULT, ZEROLOG_DEFAULT, ENABLE_DISCOVERY_DEFAULT)

	var s SmD
	var err error

	s.apiRootV2 = "/hsm/v2"
	s.serviceBaseV2 = s.apiRootV2 + "/service"
	s.valuesBaseV2 = s.serviceBaseV2 + "/values"
	s.stateBaseV2 = s.apiRootV2 + "/State"
	s.componentsBaseV2 = s.stateBaseV2 + "/Components"
	s.redfishEPBaseV2 = s.apiRootV2 + "/Inventory/RedfishEndpoints"
	s.nodeMapBaseV2 = s.apiRootV2 + "/Defaults/NodeMaps"
	s.compEPBaseV2 = s.apiRootV2 + "/Inventory/ComponentEndpoints"
	s.serviceEPBaseV2 = s.apiRootV2 + "/Inventory/ServiceEndpoints"
	s.compEthIntBaseV2 = s.apiRootV2 + "/Inventory/EthernetInterfaces"
	s.hsnIntBaseV2 = s.apiRootV2 + "/Inventory/HSNInterfaces"
	s.hwinvByLocBaseV2 = s.apiRootV2 + "/Inventory/Hardware"
	s.hwinvByFRUBaseV2 = s.apiRootV2 + "/Inventory/HardwareByFRU"
	s.invDiscoverBaseV2 = s.apiRootV2 + "/Inventory/Discover"
	s.invDiscStatusBaseV2 = s.apiRootV2 + "/Inventory/DiscoveryStatus"
	s.subscriptionBaseV2 = s.apiRootV2 + "/Subscriptions"
	s.groupsBaseV2 = s.apiRootV2 + "/groups"
	s.partitionsBaseV2 = s.apiRootV2 + "/partitions"
	s.membershipsBaseV2 = s.apiRootV2 + "/memberships"
	s.compLockBaseV2 = s.apiRootV2 + "/locks"
	s.sysInfoBaseV2 = s.apiRootV2 + "/sysinfo"
	s.powerMapBaseV2 = s.sysInfoBaseV2 + "/powermaps"

	s.parseCmdLine()

	// Set up logging for State Manager
	s.lg = log.New(os.Stdout, "", log.Lshortfile|log.LstdFlags|log.Lmicroseconds)
	if err := s.SetLogLevel(LogLevel(s.logLevelIn)); err != nil {
		os.Exit(1)
	}

	serviceName, err = base.GetServiceInstanceName()
	if err != nil {
		serviceName = "SMD"
		s.LogAlways("WARNING, can't get service/instance name, using '%s'",
			serviceName)
	}

	s.LogAlwaysStr("Starting... " + serviceName + " " + Version + " " + GitCommit + "\n")
	s.LogAlwaysStr(VersionInfo())
	// Route logs from Redfish interrogration to main smd log.
	rf.SetLogger(s.lg)
	// Route logs for sm module to main smd log
	sm.SetLogger(s.lg)

	// Load HMS base configuration file
	if err := base.InitTypes(s.hmsConfigPath); err != nil {
		s.LogAlways("Error: %s\n", err)
	}

	client := s.GetHTTPClient()

	// Skip SLS if not given a URL.
	if len(s.slsUrl) != 0 {
		s.sls = slsapi.NewSLS(s.slsUrl, client, serviceName)
	}

	// Skip HBTD if not given a URL.
	if len(s.hbtdUrl) != 0 {
		s.hbtd = hbtdapi.NewHBTD(s.hbtdUrl, client, serviceName)
	}

	// Use socks, etc. proxy when interrogating Redfish endpoints
	//	if s.proxyURL != "" {
	//		s.LogAlways("Using '%s' as proxy when interrogating Redfish.",
	//			s.proxyURL)
	//		rf.SetHTTPClientProxyURL(s.proxyURL)
	//	}
	// Generate unit test output during Redfish inventory discovery
	if s.genTestPayloads != "" {
		if err := rf.EnableGenTestingPayloads(s.genTestPayloads); err != nil {
			s.LogAlways("EnableGenTestingPayloads: Error '%s'", err)
		}
	}

	// Connect to database - DSN generated/checked during option parsing
	// per dbType, so we should always be using a valid, supported type.
	if s.dbType == dbTypePostgres {
		hmsdsLgLvl := hmsds.LOG_DEFAULT
		s.LogAlways("Connecting to data store (Postgres)...")
		s.db = hmsds.NewHMSDB_PG(s.dbDSN, s.lg)
		switch s.lgLvl {
		case LOG_DEFAULT:
			hmsdsLgLvl = hmsds.LOG_DEFAULT
		case LOG_NOTICE:
			hmsdsLgLvl = hmsds.LOG_NOTICE
		case LOG_INFO:
			hmsdsLgLvl = hmsds.LOG_INFO
		case LOG_DEBUG:
			hmsdsLgLvl = hmsds.LOG_DEBUG
		default:
			hmsdsLgLvl = hmsds.LOG_DEBUG
		}
		s.db.SetLogLevel(hmsdsLgLvl)
	}
	if applyMigrations {
		s.LogAlways("Applying all unapplied migrations")
		for {
			migrateConnection, err := pgmigrate.DBConnect(s.dbDSN)
			if err != nil {
				s.LogAlways("Error connecting to database: %s", err)
				time.Sleep(5 * time.Second)
				continue
			}
			err = pgmigrate.ApplyMigrations("/persistent_migrations", migrateConnection)
			if err != nil {
				s.LogAlways("Error applying migrations: %s", err)
				time.Sleep(5 * time.Second)
				continue
			}
			break
		}
	}
	for {
		if err := s.db.Open(); err != nil {
			s.LogAlways("DB Connection failed.  Retrying in 5 seconds")
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	if s.readVault || s.writeVault {
		for {
			var err error
			s.LogAlways("Connecting to secure store (Vault)...")
			// Start a connection to Vault
			if s.ss, err = sstorage.NewVaultAdapter(""); err != nil {
				s.LogAlways("Error: Secure Store connection failed - %s", err)
				time.Sleep(5 * time.Second)
			} else {
				s.LogAlways("Connection to secure store (Vault) succeeded")

				// Check to see if we should looks elsewhere for creds.
				vaultKeypath, ok := os.LookupEnv("VAULT_KEYPATH")
				if !ok {
					vaultKeypath = "secret/hms-creds"
				}

				s.ccs = compcreds.NewCompCredStore(vaultKeypath, s.ss)
				// Kick off thread to retry failed stores to Vault
				// go s.storeCredRetry()
				break
			}
		}
	}

	//Cert mgmt support

	hms_certs.InitInstance(nil, serviceName)

	vurl := os.Getenv("SMD_LOG_INSECURE_FAILOVER")
	if vurl != "" {
		yn, _ := strconv.ParseBool(vurl)
		if !yn {
			//Defaults to true
			hms_certs.ConfigParams.LogInsecureFailover = false
		}
	}
	vurl = os.Getenv("SMD_CA_URI")
	if vurl == "" {
		s.LogAlways("CA_URI: Not specified.")
	} else {
		s.LogAlways("CA_URI: '%s'.", vurl)
	}

	//Initialize the SCN subscription list and map
	s.scnSubs.SubscriptionList = []sm.SCNSubscription{}
	s.SCNSubscriptionRefresh()

	// Start worker pool
	//TODO: Make the worker pool size a configurable value.
	s.wp = base.NewWorkerPool(42, 10000)
	s.wp.Run()

	s.wpRFEvent = base.NewWorkerPool(1000, 10000)
	s.wpRFEvent.Run()

	// Start monitoring message bus, if configured
	if s.openchami {
		s.LogAlways("OpenCHAMI: No redfish event monitoring.")
	} else {
		s.smapCompEP = NewSyncMap(ComponentEndpointSMap(&s))
		go s.StartRFEventMonitor()
		s.LogAlways("Started redfish event monitoring.")
	}

	// Start the component lock cleanup thread
	s.CompReservationCleanup()

	// Start the Job Sync thread to pick up orphaned
	// jobs from other HSM instances.
	s.jobList = make(map[string]*Job, 0)
	s.srfpJobList = make(map[string]*Job, 0)
	s.discMap = make(map[string]int, 0)
	s.JobSync()
	if !s.enableDiscovery {
		s.DiscoverySync()
		s.DiscoveryUpdater()
	}

	// Initialize token authorization and load JWKS well-knowns from .well-known endpoint
	if s.jwksURL != "" {
		s.LogAlways("Fetching public key from server...")
		for i := 0; i <= 5; i++ {
			err = s.fetchPublicKeyFromURL(s.jwksURL)
			if err != nil {
				s.LogAlways("failed to initialize auth token: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}
			s.LogAlways("Initialized the auth token successfully.")
			break
		}
	}

	// Start serving HTTP
	var router *chi.Mux
	publicRoutes := s.generatePublicRoutes()
	protectedRoutes := s.generateProtectedRoutes()
	router = s.NewRouter(publicRoutes, protectedRoutes)

	s.LogAlways("GOMAXPROCS is: %v", runtime.GOMAXPROCS(0))
	s.LogAlways("Listening for connections at: %v", s.httpListen)
	s.LogAlways("Registered SMD protected routes: %v", protectedRoutes)
	s.LogAlways("Registered SMD public routes: %v", publicRoutes)
	err = s.setupCerts(s.tlsCert, s.tlsKey)
	if err == nil {
		err = http.ListenAndServeTLS(s.httpListen, s.tlsCert, s.tlsKey, router)
	} else {
		// This is just a fallback for testing.  There will not be a non-TLS
		// method supported in the final product.
		s.LogAlways("Warning: TLS cert or key file missing, falling back to http")
		err = http.ListenAndServe(s.httpListen, router)
	}
	s.LogAlways("HTTP server error: %s\n", err)
}
