// Copyright (C) 2016  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package main

import (
	"math"
	"testing"

	"github.com/aristanetworks/goarista/test"
	"github.com/openconfig/reference/rpc/openconfig"
)

func TestParseValue(t *testing.T) { // Because parsing JSON sucks.
	testcases := []struct {
		input    string
		expected interface{}
	}{
		{"42", int64(42)},
		{"-42", int64(-42)},
		{"42.42", float64(42.42)},
		{"-42.42", float64(-42.42)},
		{`"foo"`, nil},
		{"9223372036854775807", int64(math.MaxInt64)},
		{"-9223372036854775808", int64(math.MinInt64)},
		{"9223372036854775808", uint64(math.MaxInt64) + 1},
	}
	for i, tcase := range testcases {
		actual := parseValue(&openconfig.Update{
			Value: &openconfig.Value{
				Value: []byte(tcase.input),
			},
		})
		if d := test.Diff(tcase.expected, actual); d != "" {
			t.Errorf("#%d: %s: %#v vs %#v", i, d, tcase.expected, actual)
		}
	}
}
