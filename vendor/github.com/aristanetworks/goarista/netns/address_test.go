// Copyright (C) 2016  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package netns

import (
	"testing"
)

func TestParseAddress(t *testing.T) {
	tests := []struct {
		desc string
		arg  string
		vrf  string
		addr string
		err  bool
	}{{
		"Parse address with VRF",
		"vrf1/1.2.3.4:50",
		"ns-vrf1",
		"1.2.3.4:50",
		false,
	}, {
		"Parse address without VRF",
		"1.2.3.4:50",
		"",
		"1.2.3.4:50",
		false,
	}, {
		"Parse malformed input",
		"vrf1/1.2.3.4/24",
		"",
		"",
		true,
	}}

	for _, tt := range tests {
		vrf, addr, err := ParseAddress(tt.arg)
		if tt.err {
			if err == nil {
				t.Fatalf("%s: expected error, but got success", tt.desc)
			}
		} else {
			if err != nil {
				t.Fatalf("%s: expected success, but got error %s", tt.desc, err)
			}
			if addr != tt.addr {
				t.Fatalf("%s: expected addr %s, but got %s", tt.desc, tt.addr, addr)
			}
			if vrf != tt.vrf {
				t.Fatalf("%s: expected VRF %s, but got %s", tt.desc, tt.vrf, vrf)
			}
		}
	}
}

func TestVrfToNetNSTests(t *testing.T) {
	tests := []struct {
		desc  string
		vrf   string
		netNS string
	}{{
		"Empty VRF name",
		"",
		"",
	}, {
		"Default VRF",
		"default",
		"default",
	}, {
		"Regular VRF name",
		"cust1",
		"ns-cust1",
	}}
	for _, tt := range tests {
		netNS := VRFToNetNS(tt.vrf)
		if netNS != tt.netNS {
			t.Fatalf("%s: expected netNS %s, but got %s", tt.desc, tt.netNS,
				netNS)
		}
	}
}
