// MIT License
//
// (C) Copyright [2018-2025] Hewlett Packard Enterprise Development LP
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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/OpenCHAMI/smd/v2/pkg/sm"
	base "github.com/Cray-HPE/hms-base/v2"
)

///////////////////////////////////////////////////////////////////////////////
// Job definitions
///////////////////////////////////////////////////////////////////////////////

const (
	JTYPE_INVALID base.JobType = iota
	JTYPE_SCN
	JTYPE_RFEVENT
	JTYPE_MAX
)

var JTypeString = map[base.JobType]string{
	JTYPE_INVALID: "JTYPE_INVALID",
	JTYPE_SCN:     "JTYPE_SCN",
	JTYPE_RFEVENT: "JTYPE_RFEVENT",
	JTYPE_MAX:     "JTYPE_MAX",
}

// /////////////////////////////////////////////////////////////////////////////
// Job: JTYPE_SCN
// /////////////////////////////////////////////////////////////////////////////
type JobSCN struct {
	Status base.JobStatus
	IDs    []string
	Data   base.Component
	Err    error
	s      *SmD
	Logger *log.Logger
}

// ///////////////////////////////////////////////////////////////////////////
// Create a JTYPE_SCN job data structure.
//
// ids(in):   List of XNames to be sent in the SCN
// state(in): The state of the components in 'ids'.
// s(in):     SmD instance we are working on behalf of.
// Return:    Job data structure to be used by work Q.
// ///////////////////////////////////////////////////////////////////////////
func NewJobSCN(ids []string, data base.Component, s *SmD) base.Job {
	j := new(JobSCN)
	j.Status = base.JSTAT_DEFAULT
	j.IDs = ids
	j.Data = data
	j.s = s
	j.Logger = s.lg

	return j
}

// ///////////////////////////////////////////////////////////////////////////
// Log function for SCN job. Note that for now this is just a simple
// log call, but may be expanded in the future.
//
// format(in):  Printf-like format string.
// a(in):       Printf-like argument list.
// Return:      None.
// ///////////////////////////////////////////////////////////////////////////
func (j *JobSCN) Log(format string, a ...interface{}) {
	// Use caller's line number (depth=2)
	j.Logger.Output(2, fmt.Sprintf(format, a...))
}

// ///////////////////////////////////////////////////////////////////////////
// Return current job type.
//
// Args: None
// Return: Job type.
// ///////////////////////////////////////////////////////////////////////////
func (j *JobSCN) Type() base.JobType {
	return JTYPE_SCN
}

// ///////////////////////////////////////////////////////////////////////////
// Run a job. This is done by the worker pool when popping a job off of the
// work Q/chan.
//
// Args: None.
// Return: None.
// ///////////////////////////////////////////////////////////////////////////
func (j *JobSCN) Run() {
	var trigger string
	var triggerType int
	var waitGroup sync.WaitGroup
	scn := sm.SCNPayload{
		Components:     j.IDs,
		Enabled:        j.Data.Enabled,
		Flag:           j.Data.Flag,
		Role:           j.Data.Role,
		SubRole:        j.Data.SubRole,
		SoftwareStatus: j.Data.SwStatus,
		State:          j.Data.State,
	}
	// j.s.LogAlways("Sending SCN: %v\n", scn)
	payload, err := json.Marshal(scn)
	if err != nil {
		j.s.LogAlways("WARNING: SCN failed. Could not encode JSON: %v (%v)", err, scn)
		j.SetStatus(base.JSTAT_ERROR, err)
		return
	}
	// j.s.LogAlways("Sending SCN Payload: %v\n", string(payload))
	client := j.s.GetHTTPClient()

	// Get a the state that triggered this SCN
	if len(scn.State) != 0 {
		trigger = strings.ToLower(scn.State)
		triggerType = SCNMAP_STATE
	} else if len(scn.Role) != 0 {
		trigger = strings.ToLower(scn.Role)
		triggerType = SCNMAP_ROLE
	} else if len(scn.SubRole) != 0 {
		trigger = strings.ToLower(scn.SubRole)
		triggerType = SCNMAP_SUBROLE
	} else if len(scn.SoftwareStatus) != 0 {
		trigger = strings.ToLower(scn.SoftwareStatus)
		triggerType = SCNMAP_SWSTATUS
	} else if scn.Enabled != nil {
		trigger = "enabled"
		triggerType = SCNMAP_ENABLED
	} else {
		j.s.LogAlways("warning: Invalid SCN trigger %v", scn)
		j.SetStatus(base.JSTAT_ERROR, errors.New("invalid SCN trigger"))
		return
	}
	if j.s.scnSubMap[triggerType] == nil {
		// No subscriptions for this trigger type
		return
	}
	urlList, ok := j.s.scnSubMap[triggerType][trigger]
	if !ok {
		// No URLs to send to
		return
	}
	for _, url := range urlList {
		waitGroup.Add(1)
		go func(urlStr string) {
			defer waitGroup.Done()
			for retry := 0; retry < 3; retry++ {
				var strbody []byte
				req, rerr := http.NewRequest("POST", urlStr, bytes.NewReader(payload))
				if err != nil {
					j.s.LogAlways("WARNING: can't create an HTTP request: %v",
						rerr)
					time.Sleep(5 * time.Second)
					continue
				}
				base.SetHTTPUserAgent(req, serviceName)
				req.Header.Add("Content-Type", "application/json")
				newRequest, rerr := retryablehttp.FromRequest(req)
				if err != nil {
					j.s.LogAlways("WARNING: can't create an HTTP request: %v",
						rerr)
					time.Sleep(5 * time.Second)
					continue
				}
				rsp, err := client.Do(newRequest)
				if err != nil {
					base.DrainAndCloseResponseBody(rsp)
					j.s.LogAlways("WARNING: SCN POST failed for %s: %v", urlStr, err)
				} else {
					if rsp.Body != nil {
						strbody, _ = ioutil.ReadAll(rsp.Body)
					}
					base.DrainAndCloseResponseBody(rsp)
					if rsp.StatusCode != 200 {
						j.s.LogAlways("WARNING: An error occurred uploading SCN to %s: %s %s", urlStr, rsp.Status, string(strbody))
					} else {
						return
					}
				}
				time.Sleep(5 * time.Second)
			}
		}(url.url)
	}
	waitGroup.Wait()
}

// ///////////////////////////////////////////////////////////////////////////
// Return the current job status and error info.
//
// Args: None
// Return: Current job status, and any error info (if any).
// ///////////////////////////////////////////////////////////////////////////
func (j *JobSCN) GetStatus() (base.JobStatus, error) {
	if j.Status == base.JSTAT_ERROR {
		return j.Status, j.Err
	}
	return j.Status, nil
}

// ///////////////////////////////////////////////////////////////////////////
// Set job status.
//
// newStatus(in): Status to set job to.
// err(in):       Error info to associate with the job.
// Return:        Previous job status; nil on success, error string on error.
// ///////////////////////////////////////////////////////////////////////////
func (j *JobSCN) SetStatus(newStatus base.JobStatus, err error) (base.JobStatus, error) {
	if newStatus >= base.JSTAT_MAX {
		return j.Status, errors.New("error: Invalid Status")
	} else {
		oldStatus := j.Status
		j.Status = newStatus
		j.Err = err
		return oldStatus, nil
	}
}

// ///////////////////////////////////////////////////////////////////////////
// Cancel a job.  Note that this JobType does not support cancelling the
// job while it is being processed
//
// Args:   None
// Return: Current job status before cancelling.
// ///////////////////////////////////////////////////////////////////////////
func (j *JobSCN) Cancel() base.JobStatus {
	if j.Status == base.JSTAT_QUEUED || j.Status == base.JSTAT_DEFAULT {
		j.Status = base.JSTAT_CANCELLED
	}
	return j.Status
}

// /////////////////////////////////////////////////////////////////////////////
// Job: JTYPE_RFEVENT
// /////////////////////////////////////////////////////////////////////////////
type JobRFEvent struct {
	Status  base.JobStatus
	Payload string
	Err     error
	s       *SmD
	Logger  *log.Logger
}

// ///////////////////////////////////////////////////////////////////////////
// Create a JTYPE_RFEVENT job data structure.
//
// payload(in): The raw redfish event to process
// s(in):     SmD instance we are working on behalf of.
// Return:    Job data structure to be used by work Q.
// ///////////////////////////////////////////////////////////////////////////
func NewJobRFEvent(payload string, s *SmD) base.Job {
	j := new(JobRFEvent)
	j.Status = base.JSTAT_DEFAULT
	j.Payload = payload
	j.s = s
	j.Logger = s.lg

	return j
}

// ///////////////////////////////////////////////////////////////////////////
// Log function for SCN job. Note that for now this is just a simple
// log call, but may be expanded in the future.
//
// format(in):  Printf-like format string.
// a(in):       Printf-like argument list.
// Return:      None.
// ///////////////////////////////////////////////////////////////////////////
func (j *JobRFEvent) Log(format string, a ...interface{}) {
	// Use caller's line number (depth=2)
	j.Logger.Output(2, fmt.Sprintf(format, a...))
}

// ///////////////////////////////////////////////////////////////////////////
// Return current job type.
//
// Args: None
// Return: Job type.
// ///////////////////////////////////////////////////////////////////////////
func (j *JobRFEvent) Type() base.JobType {
	return JTYPE_RFEVENT
}

// ///////////////////////////////////////////////////////////////////////////
// Run a job. This is done by the worker pool when popping a job off of the
// work Q/chan.
//
// Args: None.
// Return: None.
// ///////////////////////////////////////////////////////////////////////////
func (j *JobRFEvent) Run() {
	err := j.s.doHandleRFEvent(j.Payload)
	if err != nil {
		j.s.Log(LOG_INFO, "Got error '%s' processing event: %s", err, j.Payload)
	}
}

// ///////////////////////////////////////////////////////////////////////////
// Return the current job status and error info.
//
// Args: None
// Return: Current job status, and any error info (if any).
// ///////////////////////////////////////////////////////////////////////////
func (j *JobRFEvent) GetStatus() (base.JobStatus, error) {
	if j.Status == base.JSTAT_ERROR {
		return j.Status, j.Err
	}
	return j.Status, nil
}

// ///////////////////////////////////////////////////////////////////////////
// Set job status.
//
// newStatus(in): Status to set job to.
// err(in):       Error info to associate with the job.
// Return:        Previous job status; nil on success, error string on error.
// ///////////////////////////////////////////////////////////////////////////
func (j *JobRFEvent) SetStatus(newStatus base.JobStatus, err error) (base.JobStatus, error) {
	if newStatus >= base.JSTAT_MAX {
		return j.Status, errors.New("error: Invalid Status")
	} else {
		oldStatus := j.Status
		j.Status = newStatus
		j.Err = err
		return oldStatus, nil
	}
}

// ///////////////////////////////////////////////////////////////////////////
// Cancel a job.  Note that this JobType does not support cancelling the
// job while it is being processed
//
// Args:   None
// Return: Current job status before cancelling.
// ///////////////////////////////////////////////////////////////////////////
func (j *JobRFEvent) Cancel() base.JobStatus {
	if j.Status == base.JSTAT_QUEUED || j.Status == base.JSTAT_DEFAULT {
		j.Status = base.JSTAT_CANCELLED
	}
	return j.Status
}
