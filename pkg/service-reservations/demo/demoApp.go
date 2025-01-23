// MIT License
//
// (C) Copyright [2020-2023] Hewlett Packard Enterprise Development LP
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
	"time"

	reservation "github.com/Cray-HPE/hms-smd/v2/pkg/service-reservations"
	"github.com/sirupsen/logrus"
)

func main() {

	//setup the reservation object
	var res reservation.Production
	logger := logrus.New()

	logger.Level = logrus.DebugLevel

	Formatter := new(logrus.TextFormatter)
	Formatter.TimestampFormat = "2006-01-02T15:04:05.999999999Z07:00"
	Formatter.FullTimestamp = true
	Formatter.ForceColors = true
	logger.SetFormatter(Formatter)
	//logger.SetReportCaller(true)

	//^^^ setup a logger of your choosing; or pass nil

	// initialize the object
	res.Init("http://localhost:27779", "/hsm/v2/locks/service/reservations", 0, logger)

	//list out a few xnames you want to reserve
	xnames := []string{"x0c0s1b0", "x0c0s2b0"}

	//aquire the reservation
	err := res.Aquire(xnames)
	if err != nil {
		logger.Error(err)
	} else {
		//IT WORKED!
		logger.Info("Aquired reservation")
	}

	//go get the status of ALL reservations
	status := res.Status()
	logrus.WithField("MAP", status).Info("Status")

	for i := 0; i < 12; i++ {
		//sleep to emulate doing other things; if logging is on, you will see it renew and check for renewal
		time.Sleep(time.Second * 8)
	}

	//When you are done, release the reservations
	err = res.Release(xnames)
	if err != nil {
		logger.Error(err)
	}

	//Lets force a few errors if we can...
	// What happens if we call relase multiple times?
	err = res.Release(xnames)
	if err != nil {
		logger.Error(err)
	}

	//Calling release multiple times WILL cause an error; but most likely its because it already was released

	// Lets aquire the reservations again
	err = res.Aquire(xnames)
	if err != nil {
		logger.Error(err)
	} else {
		//IT WORKED!
		logger.Info("Aquired reservation")
	}

	//go get the status of ALL reservations, Youll notice that you the keys have changed.
	status = res.Status()
	logrus.WithField("MAP", status).Info("Status")

	// lets force another error by trying to aquire it again
	err = res.Aquire(xnames)
	if err != nil {
		logger.Error(err)
	} else {
		//IT WORKED!
		logger.Info("Aquired reservation")
	}

	//The client has no control over how frequently the library will renew; The contract is that within 30 seconds of expiration
	//the library will send a renewal request. The library wakes up every 10 seconds to make sure it doesnt miss a period.
	//if the xname is in the status map we must trust the guarantee that the library will perform the renewals w/n a correct time.
	//therefore unless the SMD times out; or unless someone yanks the rug out from under it; the check function is all that
	//is needed to make sure we are still good to go.  b/c this is a distributed system I cannot make an absolute guarantee,
	//but it will be close enough that barring things outside our control it will work.

	//Lets CHECK the xnames
	isValid := res.Check(xnames)
	if !isValid {
		logger.Error("Xnames not valid, we dont have the reservation!")
	} else {
		logger.Debug("xnames are valid, we have the reservation")
	}

	//Lets release the reservations
	err = res.Release(xnames)
	if err != nil {
		logger.Error(err)
	}

	//Lets CHECK the xnames again
	isValid = res.Check(xnames)
	if !isValid {
		logger.Error("Xnames not valid, we dont have the reservation!")
	} else {
		logger.Debug("xnames are valid, we have the reservation")
	}

}
