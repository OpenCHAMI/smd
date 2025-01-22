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
package rf

import (
	"testing"
	"time"
)

func TestSetHTTPClientTimeout(t *testing.T) {
	SetHTTPClientTimeout(50)
	if GetHTTPClientTimeout() != 50 {
		t.Errorf("Test 1: FAIL: bad value: %d", GetHTTPClientTimeout())
	}
	SetHTTPClientTimeout(-1)
	if GetHTTPClientTimeout() != 50 {
		t.Errorf("Test 2: FAIL: bad value: %d", GetHTTPClientTimeout())
	}
}

/*
func TestSetHTTPClientProxyURL(t *testing.T) {
	SetHTTPClientProxyURL("socks5://127.0.0.1:9999")
	if GetHTTPClientProxyURL() != "socks5://127.0.0.1:9999" {
		t.Errorf("Test 1: FAIL: bad value: %s", GetHTTPClientProxyURL())
	}
}

func TestSetHTTPClientInsecureSkipVerify(t *testing.T) {
	SetHTTPClientInsecureSkipVerify(false)
	if GetHTTPClientInsecureSkipVerify() != false {
		t.Errorf("Test 1: FAIL: wrong value (true)")
	}
}
*/

func TestRfDefaultClient(t *testing.T) {
	httpRFClient = nil
	SetHTTPClientTimeout(35)
	client := RfDefaultClient()
	if client.InsecureClient.HTTPClient.Timeout != time.Duration(35)*time.Second {
		t.Errorf("Test 1: FAIL: Got unexpected/no timeout")
	}
}

/*
func TestRfProxyClient(t *testing.T) {
	SetHTTPClientTimeout(66)
	client := RfProxyClient("asdfasdfasdf")
	if client.Timeout != time.Duration(66)*time.Second {
		t.Errorf("Test 1: FAIL: Got unexpected/no timeout")
	}
	// Don't know a way to verify other fields, we get a generic
	// interface type back and can't seem to cast it.
	SetHTTPClientTimeout(77)
	client = RfProxyClient("socks5://127.0.0.1:9999")
	if client.Timeout != time.Duration(77)*time.Second {
		t.Errorf("Test 2: FAIL: Got unexpected/no timeout")
	}
}
*/
