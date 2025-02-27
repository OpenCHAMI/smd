// MIT License
//
// (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
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
	base "github.com/Cray-HPE/hms-base"
	msgbus "github.com/Cray-HPE/hms-msgbus"
	"strconv"
	"strings"
)

var MsgBusAlreadyConnected = base.NewHMSError("sm_msg",
	"operation not allowed when connection is active")
var MsgBusInvalidHostSpec = base.NewHMSError("sm_msg",
	"invalid telemetry host specification")
var MsgBusMissingHostSpec = base.NewHMSError("sm_msg",
	"missing telemetry host specification")

const RF_EVENT_MSGBUS_TOPIC = "telemetry"
const RF_EVENT_MSGBUS_GROUPID = "cray-hms-smd"

var msgbusConfigDefaults = msgbus.MsgBusConfig{
	BusTech:        msgbus.BusTechKafka,
	Blocking:       msgbus.Blocking,
	Direction:      msgbus.BusReader,
	ConnectRetries: 10,
	Topic:          RF_EVENT_MSGBUS_TOPIC,
	GroupId:        RF_EVENT_MSGBUS_GROUPID,
}

// Convenience function to parse a host:port specification.
//
// hspec(in): Host:port specification.
// Return:    Hostname; Port number; Topic; Error code on failure, or nil.
func (s *SmD) getTelemetryHost(hspec string) (string, int, string, error) {
	toks := strings.Split(hspec, ":")
	if len(toks) != 3 {
		s.LogAlways("Invalid telemetry host specification '%s', "+
			"should be host:port:topic format.",
			hspec)
		return "", 0, "", MsgBusInvalidHostSpec
	}
	port, perr := strconv.Atoi(toks[1])
	if perr != nil {
		s.LogAlways("Invalid port specification '%s', must be numeric.",
			toks[1])
		return "", 0, "", MsgBusInvalidHostSpec
	}
	return toks[0], port, toks[2], nil
}

// Set up config for message bus connection.   Should be called before
// MsgBusConnect().  Should not be called while a connection is active.
//
// hspec(in): Host:port specification.
// Error code on invalid input, or nil
func (s *SmD) MsgBusConfig(hspec string) error {
	if s.msgbusHandle != nil {
		return MsgBusAlreadyConnected
	}
	s.msgbusConfig = msgbusConfigDefaults
	host, port, topic, err := s.getTelemetryHost(hspec)
	if err != nil {
		return err
	}
	s.msgbusConfig.Host = host
	s.msgbusConfig.Port = port
	s.msgbusConfig.Topic = topic
	return nil
}

// Connects to message bus, assuming MsgBusConfig has been called.
// Returns error != nil if config is missing/bad or if a connection is
// already active on this SmD object.
func (s *SmD) MsgBusConnect() error {
	var err error
	if s.msgbusHandle != nil {
		return MsgBusAlreadyConnected
	}
	if s.msgbusConfig.Host == "" ||
		s.msgbusConfig.Port == 0 ||
		s.msgbusConfig.Topic == "" {

		return MsgBusMissingHostSpec
	}
	s.msgbusHandle, err = msgbus.Connect(s.msgbusConfig)
	if err != nil {
		s.msgbusHandle = nil
	}
	return err
}

// Disconnects from message bus if connected.  On failure, error is non-nil but
// either way, handle is set to nil and connection is still considered closed.
// This leaves the possibility of a stale old connection, but this is better
// than forcing us to try again and hoping that the disconnect will not return
// an error on a subsequent attempt, something that may never happen.
func (s *SmD) MsgBusDisconnect() error {
	if s.msgbusHandle == nil {
		return nil
	}
	err := s.msgbusHandle.Disconnect()
	s.msgbusHandle = nil
	return err
}

// Does a disconnect (a no-op if not connected) followed by a connect.
// If error is non-nil the connection did not appear to be re-established
// and reconnect should be tried again.
func (s *SmD) MsgBusReconnect() error {
	err := s.MsgBusDisconnect()
	if err != nil {
		return nil
	}
	return s.MsgBusConnect()
}

// Read a string representing a complete message read from the message bus.
// By default this is a blocking call.  Typically messages should be read
// serially by a single connection as that is the unit of parallelism.
func (s *SmD) MsgBusReadNext() (string, error) {
	return s.msgbusHandle.MessageRead()
}
