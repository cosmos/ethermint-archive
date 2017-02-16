// Copyright (C) 2016  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package netns_test

import (
	"net"
	"net/http"
	"time"

	"github.com/aristanetworks/goarista/netns"
)

func ExampleDo_httpClient() {
	vrf := "management"
	vrf = netns.VRFToNetNS(vrf) // vrf is now "ns-management"

	dial := func(network, address string) (net.Conn, error) {
		var conn net.Conn
		err := netns.Do(vrf, func() error {
			var err error
			conn, err = (&net.Dialer{
				Timeout:   30 * time.Second, // This is the connection timeout
				KeepAlive: 30 * time.Second,
			}).Dial(network, address)
			return err
		})
		return conn, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			//TLSClientConfig: ..., <- if you need SSL/TLS.
			Dial: dial,
		},
		Timeout: 30 * time.Second, // This is the request timeout
	}

	resp, err := client.Get("http://example.com")
	_ = resp
	_ = err
}
