// MIT License
//
// (C) Copyright [2020-2025] Hewlett Packard Enterprise Development LP
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

package service_reservations

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/OpenCHAMI/smd/v2/pkg/sm"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

const HSM_DEFAULT_RESERVATION_PATH = "/hsm/v2/locks/service/reservations"
const HSM_DEFAULT_SERVER = "https://api-gw-service-nmn/apis/smd"
const DEFAULT_TERM_MINUTES = 1
const DefaultExpirationWindow = 30
const DefaultSleepTime = 10

func DrainAndCloseResponseBodyAndCancelContext(resp *http.Response, ctxCancel context.CancelFunc) {
	// Must always drain and close response body first
	base.DrainAndCloseResponseBody(resp)

	// Call context cancel function, if supplied.  This must always be done
	// after draining and closing the response body
	if ctxCancel != nil {
			ctxCancel()
	}
}

//TODO - Future Enhancements
// * ALLOW both rigid an flexible processing -> not sure rigid release makes sense... so really just flexible on the AQUIRE
// * alter CHECK, RELEASE functions to cope with flexibility -> return an array of tuples: something like xname, bool, error, etc
// * consider if the aquire and release should be idempotent, such that a second call right away doesnt error; or returns a special error

//TODO - Today
// Sample application

// Implementation considerations
// Should I periodically call CloseIdleConnections? (perhaps in the renewal function)  So far dont see the need
// Should I pass a reference to the http client?
// Should I return the actual structures? At this point I dont think there is a need.
// Should I provide a way to shutdown the renewal thread? This will be either used in a service FOREVER or in a simple app; so not needed imho.

type Reservation struct {
	Xname          string
	Expiration     time.Time
	DeputyKey      string
	ReservationKey string
}

type Production struct {
	httpClient         *retryablehttp.Client
	stateManagerServer string
	reservationPath    string
	defaultTermMinutes int
	reservedMap        map[string]Reservation
	reservationMutex   sync.Mutex
	configured         bool
	logger             *logrus.Logger
}

// This uses the rigid implementation, so we will assume for EVERY operation that is all or nothing.  This could be
// pretty easily extended to support the flexible implementation, but this wasnt needed for FAS/CAPMC
type ServiceReservation interface {
	Init(stateManagerServer string, reservationPath string, defaultTermMinutes int, logger *logrus.Logger)

	InitInstance(stateManagerServer string, reservationPath string, defaultTermMinutes int, logger *logrus.Logger, svcName string)

	//Try to aquire locks for a list of xnames, renewing them within 30 seconds of expiration.
	Aquire(xnames []string) error

	//Same as Aquire() will acquire what it can, indicate which succeeded/failed
	FlexAquire(xname []string) (ReservationCreateResponse, error)

	//Restarts periodic renew for already owned reservations
	Reacquire(reservationKeys []Key, flex bool) (ReservationReleaseRenewResponse, error)

	//Validate that I still own the lock for the xnames listed.
	Check(xnames []string) bool

	//Validate which locks I still own from the xnames listed.
	FlexCheck(xnames []string) (ReservationCreateResponse, bool)

	//Validate deputy keys given by another actor
	ValidateDeputyKeys(keys []Key) (ReservationCheckResponse, error)

	//Release the locks on the xnames
	Release(xnames []string) error

	//Release from a list of xnames, only fail on total failure
	FlexRelease(xnames []string) (ReservationReleaseRenewResponse, error)

	Status() map[string]Reservation
}

var serviceName string

func (i *Production) Init(stateManagerServer string, reservationPath string, defaultTermMinutes int, logger *logrus.Logger) {

	if i.configured == false { // ONE TIME ONLY!
		i.configured = true

		if logger != nil {
			i.logger = logger
		} else {
			i.logger = logrus.New()
		}

		//setup clients and defaults

		i.httpClient = retryablehttp.NewClient()
		i.httpClient.HTTPClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		i.httpClient.RetryMax = 5
		i.httpClient.HTTPClient.Timeout = time.Second * 40
		//turn off the http client loggin!
		tmpLogger := logrus.New()
		tmpLogger.SetLevel(logrus.PanicLevel)
		i.httpClient.Logger = tmpLogger

		if len(stateManagerServer) == 0 {
			i.stateManagerServer = HSM_DEFAULT_SERVER
		} else {
			i.stateManagerServer = stateManagerServer
		}

		if len(reservationPath) == 0 {
			i.reservationPath = HSM_DEFAULT_RESERVATION_PATH
		} else {
			i.reservationPath = reservationPath
		}

		if defaultTermMinutes < 1 || defaultTermMinutes > 15 {
			i.defaultTermMinutes = DEFAULT_TERM_MINUTES
		} else {
			i.defaultTermMinutes = defaultTermMinutes
		}

		i.reservationMutex = sync.Mutex{}

		i.reservedMap = make(map[string]Reservation)

		go i.doRenewal()
	}
}

func (i *Production) InitInstance(stateManagerServer string, reservationPath string, defaultTermMinutes int, logger *logrus.Logger, svcName string) {
	serviceName = svcName
	i.Init(stateManagerServer, reservationPath, defaultTermMinutes, logger)
}

// Lets make this really simple; Im going to wake up and see what expires in the next 30 seconds;
// then I will renew those things
func (i *Production) doRenewal() {

	for ; ; time.Sleep(time.Duration(DefaultSleepTime) * time.Second) {
		i.logger.Trace("doRenewal() - TOP Loop")

		i.update() // once per loop call the update function to see what is still viable!

		if len(i.reservedMap) == 0 {
			i.logger.Debug("doRenewal() - CONTINUE - empty map, nothing to do")
			continue
		}

		var renewalCandidates []string
		for k, v := range i.reservedMap {
			if v.Expiration.Before(time.Now()) { //expired -> no point in renewing!
				i.logger.WithFields(logrus.Fields{"reservation": v, "expiration time": v.Expiration, "clock time": time.Now()}).Trace("doRenewal() - CONTINUE - Deleting from map @1")

				i.reservationMutex.Lock()
				delete(i.reservedMap, k)
				i.reservationMutex.Unlock()

				continue
			} else if v.Expiration.Add(DefaultExpirationWindow * -1 * time.Second).Before(time.Now()) {
				// is it within the 30 second expiration window?
				renewalCandidates = append(renewalCandidates, k)
			}
		}

		if len(renewalCandidates) == 0 {
			i.logger.Debug("doRenewal() - CONTINUE - Nothing to renew")
			continue
		}

		var resKeys []Key

		for _, xname := range renewalCandidates {
			if res, ok := i.reservedMap[xname]; ok {

				key := Key{
					ID:  res.Xname,
					Key: res.ReservationKey,
				}

				//add the key to the list
				resKeys = append(resKeys, key)
			}
		}

		response, err := i.doRenew(resKeys, true)
		if err != nil {
			i.logger.WithField("error", err).Error("doRenewal() - CONTINUE")
			continue
		}

		i.logger.WithFields(logrus.Fields{"Total": response.Counts.Total,
			"Success": response.Counts.Success,
			"Failure": response.Counts.Failure}).Debug("doRenewal() - renewal action complete")
		i.logger.WithFields(logrus.Fields{"Total": response.Counts.Total,
			"Success": response.Counts.Success,
			"Failure": response.Counts.Failure}).Info("ServiceReservations - renewal action complete")

		// Remove the failed ones
		for _, v := range response.Failure {
			if _, ok := i.reservedMap[v.ID]; ok {
				i.logger.WithFields(logrus.Fields{"reservation": i.reservedMap[v.ID]}).Trace("doRenewal() - deleting failures")
				i.reservationMutex.Lock()
				delete(i.reservedMap, v.ID)
				i.reservationMutex.Unlock()
			}
		}

		i.logger.Trace("doRenewal() - BOTTOM Loop")
	}
}

// Send a renew request to HSM
func (i *Production) doRenew(reservationKeys []Key, flex bool) (ReservationReleaseRenewResponse, error) {
	var response ReservationReleaseRenewResponse

	i.logger.Trace("doRenew() - START")

	processingModel := CLProcessingModelFlex
	if !flex {
		processingModel = CLProcessingModelRigid
	}

	renewalParameters := ReservationRenewalParameters{
		ProcessingModel:     processingModel,
		ReservationDuration: i.defaultTermMinutes,
		ReservationKeys:     reservationKeys,
	}

	if len(renewalParameters.ReservationKeys) == 0 {
		i.logger.Debug("doRenew() - END - Nothing to renew")
		return response, fmt.Errorf("Nothing to renew")
	}

	marshalReleaseParams, _ := json.Marshal(renewalParameters)
	stringReleaseParams := string(marshalReleaseParams)
	targetURL, _ := url.Parse(i.stateManagerServer + i.reservationPath + "/renew")

	i.logger.WithField("params", stringReleaseParams).Trace("doRenew() - Sending command")

	newRequest, err := http.NewRequest("POST", targetURL.String(), bytes.NewBuffer([]byte(stringReleaseParams)))
	if err != nil {
		i.logger.WithField("error", err).Error("doRenew() - END")
		return response, err
	}
	base.SetHTTPUserAgent(newRequest, serviceName)

	reqContext, reqCtxCancel := context.WithTimeout(context.Background(), time.Second*40)
	req, err := retryablehttp.FromRequest(newRequest)
	req = req.WithContext(reqContext)
	if err != nil {
		reqCtxCancel()
		i.logger.WithField("error", err).Error("doRenew() - END")
		return response, err
	}

	req.Header.Add("Content-Type", "application/json")

	//make request
	resp, err := i.httpClient.Do(req)
	defer DrainAndCloseResponseBodyAndCancelContext(resp, reqCtxCancel)
	if err != nil {
		i.logger.WithField("error", err).Error("doRenew() - END")
		return response, err
	}

	//process response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		i.logger.WithField("error", err).Error("doRenew() - END")
		return response, err
	}

	i.logger.WithField("response", string(body)).Trace("doRenew() - Received response")

	switch statusCode := resp.StatusCode; statusCode {

	case http.StatusOK:
		err = json.Unmarshal(body, &response)
		if err != nil {
			i.logger.WithField("error", err).Error("doRenew() - END")
			return response, err
		}
		i.logger.Trace("doRenew() - END")
		return response, nil

	case http.StatusBadRequest:
		var errResponse Problem7807
		_ = json.Unmarshal(body, &errResponse)
		err = errors.New(errResponse.Detail)
		i.logger.WithField("error", err).Error("doRenew() - END")
		return response, err

	default:
		i.logger.WithFields(logrus.Fields{"error": "failed to renew"}).Error("doRenew() - END")
		return response, fmt.Errorf("failed to renew")
	}
}

func (i *Production) Aquire(xnames []string) error {
	i.logger.Trace("Aquire() - START")

	//prepare the request
	reservation := ReservationCreateParameters{
		ID:                  xnames,
		ProcessingModel:     CLProcessingModelRigid,
		ReservationDuration: i.defaultTermMinutes,
	}
	marshalReservation, _ := json.Marshal(reservation)
	stringReservation := string(marshalReservation)
	targetURL, _ := url.Parse(i.stateManagerServer + i.reservationPath)
	i.logger.WithField("params", stringReservation).Trace("Aquire() - Sending command")
	newRequest, err := http.NewRequest("POST", targetURL.String(), bytes.NewBuffer([]byte(stringReservation)))
	if err != nil {
		i.logger.WithField("error", err).Error("Aquire() - END")
		return err
	}
	base.SetHTTPUserAgent(newRequest, serviceName)

	reqContext, reqCtxCancel := context.WithTimeout(context.Background(), time.Second*40)
	req, err := retryablehttp.FromRequest(newRequest)
	req = req.WithContext(reqContext)
	if err != nil {
		reqCtxCancel()
		i.logger.WithField("error", err).Error("Aquire() - END")
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	//make request
	resp, err := i.httpClient.Do(req)
	defer DrainAndCloseResponseBodyAndCancelContext(resp, reqCtxCancel)
	if err != nil {
		i.logger.WithField("error", err).Error("Aquire() - END")
		return err
	}

	//process response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		i.logger.WithField("error", err).Error("Aquire() - END")
		return err
	}

	i.logger.WithField("response", string(body)).Trace("Aquire() - Recieved response")

	switch statusCode := resp.StatusCode; statusCode {

	case http.StatusOK:
		var response ReservationCreateResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			i.logger.WithField("error", err).Error("Aquire() - END")
			return err
		}

		if len(response.Success) < len(xnames) {
			// because we have HARDCODED rigid; if failure > 0 or successes < xnames,
			// success MUST == 0
			err = errors.New("at least one xname could not be reserved")
			i.logger.WithField("error", err).Error("Aquire() - END")
			return err
		}

		//else load the maps
		for _, v := range response.Success {
			res := Reservation{
				Xname:          v.ID,
				DeputyKey:      v.DeputyKey,
				ReservationKey: v.ReservationKey,
			}
			res.Expiration, _ = time.Parse(time.RFC3339, v.ExpirationTime)
			i.reservationMutex.Lock()
			i.reservedMap[res.Xname] = res
			i.reservationMutex.Unlock()
		}
		i.logger.Trace("Aquire() - END")
		return nil

	case http.StatusBadRequest:
		var response Problem7807
		_ = json.Unmarshal(body, &response)
		err = errors.New(response.Detail)
		i.logger.WithField("error", err).Error("Aquire() - END")
		return err

	default:
		err = errors.New(string(body))
		i.logger.WithField("error", err).Error("Aquire() - END")
		return err
	}
}

func (i *Production) FlexAquire(xnames []string) (ReservationCreateResponse, error) {
	var resResponse ReservationCreateResponse

	i.logger.Trace("SoftAquire() - START")

	//prepare the request
	reservation := ReservationCreateParameters{
		ID:                  xnames,
		ProcessingModel:     CLProcessingModelFlex,
		ReservationDuration: i.defaultTermMinutes,
	}
	marshalReservation, _ := json.Marshal(reservation)
	stringReservation := string(marshalReservation)
	targetURL, _ := url.Parse(i.stateManagerServer + i.reservationPath)
	i.logger.WithField("params", stringReservation).Trace("SoftAquire() - Sending command")
	newRequest, err := http.NewRequest("POST", targetURL.String(), bytes.NewBuffer([]byte(stringReservation)))
	if err != nil {
		i.logger.WithField("error", err).Error("SoftAquire() - END")
		return resResponse, err
	}
	base.SetHTTPUserAgent(newRequest, serviceName)

	reqContext, reqCtxCancel := context.WithTimeout(context.Background(), time.Second*40)
	req, err := retryablehttp.FromRequest(newRequest)
	req = req.WithContext(reqContext)
	if err != nil {
		reqCtxCancel()
		i.logger.WithField("error", err).Error("SoftAquire() - END")
		return resResponse, err
	}

	req.Header.Add("Content-Type", "application/json")

	//make request
	resp, err := i.httpClient.Do(req)
	defer DrainAndCloseResponseBodyAndCancelContext(resp, reqCtxCancel)
	if err != nil {
		i.logger.WithField("error", err).Error("SoftAquire() - END")
		return resResponse, err
	}

	//process response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		i.logger.WithField("error", err).Error("SoftAquire() - END")
		return resResponse, err
	}

	i.logger.WithField("response", string(body)).Trace("SoftAquire() - Recieved response")

	switch statusCode := resp.StatusCode; statusCode {

	case http.StatusOK:
		err = json.Unmarshal(body, &resResponse)
		if err != nil {
			i.logger.WithField("error", err).Error("SoftAquire() - END")
			return resResponse, err
		}

		for _, v := range resResponse.Success {
			res := Reservation{
				Xname:          v.ID,
				DeputyKey:      v.DeputyKey,
				ReservationKey: v.ReservationKey,
			}
			res.Expiration, _ = time.Parse(time.RFC3339, v.ExpirationTime)
			i.reservationMutex.Lock()
			i.reservedMap[res.Xname] = res
			i.reservationMutex.Unlock()
		}
		i.logger.Trace("SoftAquire() - END")

	case http.StatusBadRequest:
		var pResponse Problem7807
		_ = json.Unmarshal(body, &pResponse)
		err = errors.New(pResponse.Detail)
		i.logger.WithField("error", err).Error("SoftAquire() - END")
		return resResponse, err

	default:
		err = errors.New(string(body))
		i.logger.WithField("error", err).Error("SoftAquire() - END")
		return resResponse, err
	}

	return resResponse, nil
}

// Restarts periodic renew for already owned reservations
func (i *Production) Reacquire(reservations []Reservation, flex bool) (ReservationReleaseRenewResponse, error) {
	var (
		response ReservationReleaseRenewResponse
		resKeys  []Key
	)

	i.logger.Trace("Reacquire() - START")

	resMap := make(map[string]Reservation)

	for _, res := range reservations {
		resMap[res.Xname] = res
		key := Key{
			ID:  res.Xname,
			Key: res.ReservationKey,
		}
		resKeys = append(resKeys, key)
	}
	if len(resKeys) == 0 {
		i.logger.Debug("Reacquire() - END - Nothing to reacquire")
		return response, fmt.Errorf("Nothing to reacquire")
	}
	response, err := i.doRenew(resKeys, flex)
	if err != nil {
		i.logger.WithField("error", err).Error("Reacquire() - END")
		return response, err
	}

	// Add the successfully renewed components
	for _, id := range response.Success.ComponentIDs {
		if res, ok := resMap[id]; ok {
			i.reservationMutex.Lock()
			i.reservedMap[res.Xname] = res
			i.reservationMutex.Unlock()
		}
	}
	i.logger.Trace("Reacquire() - END")
	return response, nil
}

func (i *Production) Check(xnames []string) bool {
	i.logger.Trace("Check() - START")

	valid := true
	for _, xname := range xnames {
		if _, ok := i.reservedMap[xname]; !ok {
			valid = false
			break
		}
	}
	i.logger.Trace("Check() - END")
	return valid
}

func (i *Production) FlexCheck(xnames []string) (ReservationCreateResponse, bool) {
	var retData ReservationCreateResponse
	i.logger.Trace("SoftCheck() - START")

	valid := true
	for _, xname := range xnames {
		if comp, ok := i.reservedMap[xname]; ok {
			retData.Success = append(retData.Success,
				ReservationCreateSuccessResponse{ID: comp.Xname,
					DeputyKey:      comp.DeputyKey,
					ReservationKey: comp.ReservationKey,
					ExpirationTime: comp.Expiration.Format(time.RFC3339)})
		} else {
			i.logger.Tracef("SoftCheck() - no reservation match for '%s'", xname)
			retData.Failure = append(retData.Failure,
				FailureResponse{ID: xname, Reason: "Reservation not found."})
			valid = false
		}
	}

	return retData, valid
}

func (i *Production) Release(xnames []string) error {

	i.logger.Trace("Release() - START")
	if len(xnames) == 0 {
		i.logger.Trace("release() - END -> nothing to do ")
		err := errors.New("empty set; failing release operation")
		return err
	}

	var releaseParams ReservationReleaseParameters
	releaseParams.ProcessingModel = "flexible" //TODO should this be rigid?

	for _, xname := range xnames {
		if res, ok := i.reservedMap[xname]; ok {

			key := Key{
				ID:  res.Xname,
				Key: res.ReservationKey,
			}

			//add the key to the list
			releaseParams.ReservationKeys = append(releaseParams.ReservationKeys, key)

		} else { // xname not in map
			err := errors.New(xname + " not present; failing release operation")
			i.logger.WithField("error", err).Error("Release() - END")
			return err
		}
	}

	marshalReleaseParams, _ := json.Marshal(releaseParams)
	targetURL, _ := url.Parse(i.stateManagerServer + i.reservationPath + "/release")

	i.logger.WithField("Params:", string(marshalReleaseParams)).Trace("Release() - Sending command")

	newRequest, err := http.NewRequest("POST", targetURL.String(), bytes.NewBuffer(marshalReleaseParams))
	if err != nil {
		i.logger.WithField("error", err).Error("Release() - END")
		return err
	}
	base.SetHTTPUserAgent(newRequest, serviceName)

	reqContext, reqCtxCancel := context.WithTimeout(context.Background(), time.Second*40)
	req, err := retryablehttp.FromRequest(newRequest)
	req = req.WithContext(reqContext)
	if err != nil {
		reqCtxCancel()
		i.logger.WithField("error", err).Error("Release() - END")
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	//make request
	resp, err := i.httpClient.Do(req)
	defer DrainAndCloseResponseBodyAndCancelContext(resp, reqCtxCancel)
	if err != nil {
		i.logger.WithField("error", err).Error("Release() - END")
		return err
	}

	//process response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		i.logger.WithField("error", err).Error("Release() - END")
		return err
	}
	i.logger.WithField("response", string(body)).Trace("Release() - Received response")

	switch statusCode := resp.StatusCode; statusCode {

	case http.StatusOK:
		var response ReservationReleaseRenewResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			i.logger.WithField("error", err).Error("Release() - END")
			return err
		}

		i.logger.WithFields(logrus.Fields{"Total": response.Counts.Total,
			"Success": response.Counts.Success,
			"Failure": response.Counts.Failure}).Debug("Release() - release action complete")

		for _, xname := range response.Success.ComponentIDs {
			if _, ok := i.reservedMap[xname]; ok {
				i.logger.WithFields(logrus.Fields{"reservation": i.reservedMap[xname]}).Trace("Release() - deleting released reservations")

				i.reservationMutex.Lock()
				delete(i.reservedMap, xname)
				i.reservationMutex.Unlock()
			}
		}

		if len(response.Failure) > 0 {
			err = errors.New("at least one xname could not be released")
			i.logger.WithField("error", err).Error("Release() - END")
			return err
		}

		return nil

	case http.StatusBadRequest:
		var response Problem7807
		_ = json.Unmarshal(body, &response)
		err = errors.New(response.Detail)
		i.logger.WithField("error", err).Error("Release() - END")
		return err

	default:
		err = errors.New(string(body))
		i.logger.WithField("error", err).Error("Release() - END")
		return err
	}
}

func (i *Production) FlexRelease(xnames []string) (ReservationReleaseRenewResponse, error) {
	var retData ReservationReleaseRenewResponse

	i.logger.Trace("SoftRelease() - START")
	if len(xnames) == 0 {
		i.logger.Trace("SoftRelease() - END -> nothing to do ")
		err := errors.New("empty set; failing release operation")
		return retData, err
	}

	var releaseParams ReservationReleaseParameters
	var unMapped []FailureResponse
	releaseParams.ProcessingModel = "flexible"

	for _, xname := range xnames {
		if res, ok := i.reservedMap[xname]; ok {

			key := Key{
				ID:  res.Xname,
				Key: res.ReservationKey,
			}

			//add the key to the list
			releaseParams.ReservationKeys = append(releaseParams.ReservationKeys, key)

		} else { // xname not in map
			retData.Failure = append(retData.Failure, FailureResponse{ID: xname,
				Reason: "Reservation not found."})
			err := errors.New(xname + " not found in reservation map")
			i.logger.WithField("error", err).Error("SoftRelease() - END")
		}
	}

	marshalReleaseParams, _ := json.Marshal(releaseParams)
	targetURL, _ := url.Parse(i.stateManagerServer + i.reservationPath + "/release")

	i.logger.WithField("Params:", string(marshalReleaseParams)).Trace("SoftRelease() - Sending command")

	newRequest, err := http.NewRequest("POST", targetURL.String(), bytes.NewBuffer(marshalReleaseParams))
	if err != nil {
		i.logger.WithField("error", err).Error("SoftRelease() - END")
		return retData, err
	}
	base.SetHTTPUserAgent(newRequest, serviceName)

	reqContext, reqCtxCancel := context.WithTimeout(context.Background(), time.Second*40)
	req, err := retryablehttp.FromRequest(newRequest)
	req = req.WithContext(reqContext)
	if err != nil {
		reqCtxCancel()
		i.logger.WithField("error", err).Error("SoftRelease() - END")
		return retData, err
	}

	req.Header.Add("Content-Type", "application/json")

	//make request
	resp, err := i.httpClient.Do(req)
	defer DrainAndCloseResponseBodyAndCancelContext(resp, reqCtxCancel)
	if err != nil {
		i.logger.WithField("error", err).Error("SoftRelease() - END")
		return retData, err
	}

	//process response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		i.logger.WithField("error", err).Error("SoftRelease() - END")
		return retData, err
	}
	i.logger.WithField("response", string(body)).Trace("SoftRelease() - Received response")

	switch statusCode := resp.StatusCode; statusCode {

	case http.StatusOK:
		var response ReservationReleaseRenewResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			i.logger.WithField("error", err).Error("SoftRelease() - END")
			return retData, err
		}

		i.logger.WithFields(logrus.Fields{"Total": response.Counts.Total,
			"Success": response.Counts.Success,
			"Failure": response.Counts.Failure}).Debug("SoftRelease() - release action complete")

		for _, xname := range response.Success.ComponentIDs {
			if _, ok := i.reservedMap[xname]; ok {
				i.logger.WithFields(logrus.Fields{"reservation": i.reservedMap[xname]}).Trace("SoftRelease() - deleting released reservations")

				i.reservationMutex.Lock()
				delete(i.reservedMap, xname)
				i.reservationMutex.Unlock()
				retData.Success.ComponentIDs = append(retData.Success.ComponentIDs, xname)
			}
		}
		retData.Failure = append(retData.Failure, unMapped...)

		retData.Counts.Success = len(retData.Success.ComponentIDs)
		retData.Counts.Failure = len(retData.Failure)
		retData.Counts.Total = retData.Counts.Success + retData.Counts.Failure

	case http.StatusBadRequest:
		var response Problem7807
		_ = json.Unmarshal(body, &response)
		err = errors.New(response.Detail)
		i.logger.WithField("error", err).Error("SoftRelease() - END")
		return retData, err

	default:
		err = errors.New(string(body))
		i.logger.WithField("error", err).Error("SoftRelease() - END")
		return retData, err
	}

	return retData, nil
}

func (i *Production) update() {
	i.logger.Trace("update() - START")
	var checkParameters ReservationCheckParameters

	for _, res := range i.reservedMap {
		key := Key{
			ID:  res.Xname,
			Key: res.DeputyKey,
		}
		//add the key to the list
		checkParameters.DeputyKeys = append(checkParameters.DeputyKeys, key)
	}

	if len(checkParameters.DeputyKeys) == 0 {
		i.logger.Trace("update() - END -> nothing to do ")
		return
	}
	marshalParams, _ := json.Marshal(checkParameters)
	stringParams := string(marshalParams)
	targetURL, _ := url.Parse(i.stateManagerServer + i.reservationPath + "/check")
	i.logger.WithField("params", stringParams).Trace("update() - Sending comand")

	newRequest, err := http.NewRequest("POST", targetURL.String(), bytes.NewBuffer([]byte(stringParams)))
	if err != nil {
		//die silently
		i.logger.WithField("error", err).Error("update() - END")
		return
	}
	base.SetHTTPUserAgent(newRequest, serviceName)

	reqContext, reqCtxCancel := context.WithTimeout(context.Background(), time.Second*40)
	req, err := retryablehttp.FromRequest(newRequest)
	req = req.WithContext(reqContext)
	if err != nil {
		reqCtxCancel()
		i.logger.WithField("error", err).Error("update() - END")
		return
	}

	req.Header.Add("Content-Type", "application/json")

	//make request
	resp, err := i.httpClient.Do(req)
	defer DrainAndCloseResponseBodyAndCancelContext(resp, reqCtxCancel)
	if err != nil {
		i.logger.WithField("error", err).Error("update() - END")
		return
	}

	//process response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		i.logger.WithField("error", err).Error("update() - END")
		return
	}

	i.logger.WithField("response", string(body)).Trace("update() - Received response")

	switch statusCode := resp.StatusCode; statusCode {

	case http.StatusOK:
		var response ReservationCheckResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			i.logger.WithField("error", err).Error("update() - END")
			return
		}

		for _, v := range response.Failure {
			i.logger.WithFields(logrus.Fields{"reservation": i.reservedMap[v.ID]}).Trace("deleting: removing failures in update()")

			i.reservationMutex.Lock()
			delete(i.reservedMap, v.ID)
			i.reservationMutex.Unlock()
		}

		for _, v := range response.Success {
			if _, ok := i.reservedMap[v.ID]; ok {
				i.reservationMutex.Lock()
				tmp := i.reservedMap[v.ID]
				tmp.Expiration, _ = time.Parse(time.RFC3339, v.ExpirationTime)
				i.reservedMap[v.ID] = tmp
				i.reservationMutex.Unlock()
			} else {
				i.logger.WithFields(logrus.Fields{"reservation": i.reservedMap[v.ID]}).Trace("deleting: removing not found in update()")
				i.reservationMutex.Lock()
				delete(i.reservedMap, v.ID)
				i.reservationMutex.Unlock()
			}
		}
		return
	default:
		i.logger.Warn("failed check from update()")
		return
	}
}

func (i *Production) Status() (res map[string]Reservation) {
	i.logger.Trace("Status() - START")
	res = make(map[string]Reservation)
	for k, v := range i.reservedMap {
		res[k] = v
	}
	i.logger.Trace("Status() - END")
	return res
}

func (i *Production) ValidateDeputyKeys(keys []Key) (ReservationCheckResponse, error) {
	var lockArray sm.CompLockV2DeputyKeyArray
	var jdata sm.CompLockV2ReservationResult
	var retData ReservationCheckResponse

	rmap := make(map[string]bool)

	for _, comp := range keys {
		rmap[comp.ID] = true
		if comp.Key != "" {
			lockArray.DeputyKeys = append(lockArray.DeputyKeys,
				sm.CompLockV2Key{ID: comp.ID, Key: comp.Key})
		}
	}

	ba, baerr := json.Marshal(&lockArray)
	if baerr != nil {
		return retData, fmt.Errorf("Error marshalling deputy key array: %v", baerr)
	}

	baseURL := i.stateManagerServer + i.reservationPath + "/check"
	newRequest, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewBuffer(ba))
	if err != nil {
		return retData, fmt.Errorf("Error constructing HTTP request for deputy key check: %v",
			err)
	}
	req, err := retryablehttp.FromRequest(newRequest)
	reqContext, reqCtxCancel := context.WithTimeout(context.Background(), 40*time.Second)
	req = req.WithContext(reqContext)

	rsp, rsperr := i.httpClient.Do(req)
	defer DrainAndCloseResponseBodyAndCancelContext(rsp, reqCtxCancel)
	if (rsperr != nil) {
		return retData, fmt.Errorf("Error sending http request for deputy key check: %v",
			rsperr)
	}

	switch statusCode := rsp.StatusCode; statusCode {
	case http.StatusOK:
		body, bderr := ioutil.ReadAll(rsp.Body)
		if bderr != nil {
			return retData, fmt.Errorf("Error reading http response for deputy key check: %v",
				bderr)
		}

		err = json.Unmarshal(body, &jdata)
		if err != nil {
			return retData, fmt.Errorf("Error unmarshalling response for deputy key check: %v",
				err)
		}

		//Populate success/failure

		for _, comp := range jdata.Success {
			delete(rmap, comp.ID)
			retData.Success = append(retData.Success,
				ReservationCheckSuccessResponse{ID: comp.ID,
					DeputyKey:      comp.DeputyKey,
					ExpirationTime: comp.ExpirationTime})
		}

		for _, comp := range jdata.Failure {
			delete(rmap, comp.ID)
			retData.Failure = append(retData.Failure,
				FailureResponse{ID: comp.ID, Reason: comp.Reason})
		}

		//Any remaining map entries mean the key was not found or was invalid.

		for k, _ := range rmap {
			retData.Failure = append(retData.Failure,
				FailureResponse{ID: k, Reason: "Key not found, invalid, or expired"})
		}

	default:
		return retData, fmt.Errorf("Error response from deputy key check: %d",
			statusCode)
	}
	return retData, nil
}
