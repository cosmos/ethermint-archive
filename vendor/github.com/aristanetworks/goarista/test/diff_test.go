// Copyright (C) 2015  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package test

import (
	"testing"
)

func TestDiff(t *testing.T) {
	saved := prettyPrintDepth
	prettyPrintDepth = 4
	testcases := getDeepEqualTests(t)

	for _, test := range testcases {
		diff := Diff(test.a, test.b)
		if test.diff != diff {
			t.Errorf("Diff returned different diff\n"+
				"Diff    : %q\nExpected: %q\nFor %#v == %#v",
				diff, test.diff, test.a, test.b)
		}
	}
	prettyPrintDepth = saved
}

var benchEqual = map[string]interface{}{
	"foo": "bar",
	"bar": map[string]interface{}{
		"foo": "bar",
		"bar": map[string]interface{}{
			"foo": "bar",
		},
		"foo2": []uint32{1, 2, 5, 78, 23, 236, 346, 3456},
	},
}

var benchDiff = map[string]interface{}{
	"foo": "bar",
	"bar": map[string]interface{}{
		"foo": "bar",
		"bar": map[string]interface{}{
			"foo": "bar",
		},
		"foo2": []uint32{1, 2, 5, 78, 23, 236, 346, 3457},
	},
}

func BenchmarkEqualDeepEqual(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DeepEqual(benchEqual, benchEqual)
	}
}

func BenchmarkEqualDiff(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Diff(benchEqual, benchEqual)
	}
}

func BenchmarkDiffDeepEqual(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DeepEqual(benchEqual, benchDiff)
	}
}

func BenchmarkDiffDiff(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Diff(benchEqual, benchDiff)
	}
}
