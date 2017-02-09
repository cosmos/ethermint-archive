// Copyright (C) 2016  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package test_test

import (
	"fmt"
	"testing"
	"unsafe"

	. "github.com/aristanetworks/goarista/test"
)

type alias int

type privateByteSlice struct {
	exportme []byte
}

func TestPrettyPrint(t *testing.T) {
	// This test doesn't need to cover all the types of input as a number of
	// them are covered by other tests in this package.
	ch := make(chan int, 42)
	testcases := []struct {
		input  interface{}
		pretty string
	}{
		{true, "true"},
		{(chan int)(nil), "(chan int)(nil)"},
		{ch, fmt.Sprintf("(chan int)(%p)[42]", ch)},
		{func() {}, "func(...)"},
		{unsafe.Pointer(nil), "(unsafe.Pointer)(nil)"},
		{unsafe.Pointer(t), fmt.Sprintf("(unsafe.Pointer)(%p)", t)},
		{[]byte(nil), `[]byte(nil)`},
		{[]byte{42, 0, 42}, `[]byte("*\x00*")`},
		{[]int{42, 51}, "[]int{42, 51}"},
		{[2]int{42, 51}, "[2]int{42, 51}"},
		{[2]byte{42, 51}, "[2]uint8{42, 51}"}, // Yeah, in Go `byte' is really just `uint8'.
		{alias(42), "alias(42)"},
		{privateByteSlice{[]byte("a")}, `test_test.privateByteSlice{exportme:[]byte("a")}`},
	}
	for i, tcase := range testcases {
		actual := PrettyPrint(tcase.input)
		if actual != tcase.pretty {
			t.Errorf("#%d: Wanted %q but got %q", i, actual, tcase.pretty)
		}
	}
}
