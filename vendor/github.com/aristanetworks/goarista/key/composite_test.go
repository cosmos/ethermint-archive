// Copyright (C) 2016  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package key_test

import (
	"testing"

	. "github.com/aristanetworks/goarista/key"
	"github.com/aristanetworks/goarista/test"
)

type unhashable struct {
	f func()
	u uintptr
}

func TestBadComposite(t *testing.T) {
	test.ShouldPanicWith(t, "use of unhashable type in a map", func() {
		m := map[interface{}]struct{}{
			unhashable{func() {}, 0x42}: struct{}{},
		}
		// Use Key here to make sure init() is called.
		if _, ok := m[New("foo")]; ok {
			t.Fatal("WTF")
		}
	})
	test.ShouldPanicWith(t, "use of uncomparable type on the lhs of ==", func() {
		var a interface{}
		var b interface{}
		a = unhashable{func() {}, 0x42}
		b = unhashable{func() {}, 0x42}
		// Use Key here to make sure init() is called.
		if a == b {
			t.Fatal("WTF")
		}
	})
}
