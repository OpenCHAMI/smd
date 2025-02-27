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
	"time"
)

// This is the main thread that monitors the event bus
func (s *SmD) StartRFEventMonitor() {
	if s.msgbusListen != "" {
		if err := s.MsgBusConfig(s.msgbusListen); err != nil {
			s.LogAlways("WARNING: Cannot parse message bus host: %s", err)
		} else {
			// Loop forever trying to connect or reconnect if an error
			// occurs.
			for {
				if err := s.MsgBusReconnect(); err != nil {
					s.LogAlways("ERROR: Cannot connect to  message bus host: %s",
						err)
					s.LogAlways("Retrying msg bus connection in 5 seconds")
					time.Sleep(5 * time.Second)
					continue
				}
				s.LogAlways("Connected to message bus: %s:%d:%s",
					s.msgbusConfig.Host, s.msgbusConfig.Port, s.msgbusConfig.Topic)

				// Consume events from topic on message bus.  If an
				// error occurs, report it and reconnect.
				err := s.rfEventMonitor()
				s.LogAlways("ERROR: Event monitor returned '%s', reconnecting",
					err)
			}
		}
	} else {
		s.LogAlways("No message bus host given.")
	}
	s.LogAlways("Not listening for events on the message bus.")
	return
}

// Dequeue a single event and hand it off for processing.
func (s *SmD) rfEventMonitor() error {
	// Read the event
	s.LogAlways("Starting event monitor...")
	var err error
	for {
		// Read the event
		payload, err := s.MsgBusReadNext()
		if err != nil {
			s.Log(LOG_INFO, "Got error reading event: %s", err)
			return err
		}
		rfEventJob := NewJobRFEvent(payload, s)
		s.wpRFEvent.Queue(rfEventJob)
	}
	// Shouldn't happen
	return err
}
