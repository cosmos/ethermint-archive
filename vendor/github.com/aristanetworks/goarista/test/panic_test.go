// Copyright (C) 2015  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package test

import (
	"testing"
)

func TestShouldPanic(t *testing.T) {
	fn := func() { panic("Here we are") }

	ShouldPanic(t, fn)
}

func TestShouldPanicWithString(t *testing.T) {
	fn := func() { panic("Here we are") }

	ShouldPanicWith(t, "Here we are", fn)
}

func TestShouldPanicWithInt(t *testing.T) {
	fn := func() { panic(42) }

	ShouldPanicWith(t, 42, fn)
}

func TestShouldPanicWithStruct(t *testing.T) {
	fn := func() { panic(struct{ foo string }{foo: "panic"}) }

	ShouldPanicWith(t, struct{ foo string }{foo: "panic"}, fn)
}
