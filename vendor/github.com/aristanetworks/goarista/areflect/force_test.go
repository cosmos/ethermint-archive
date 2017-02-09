// Copyright (C) 2016  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package areflect

import (
	"reflect"
	"testing"
)

type somestruct struct {
	a uint32
}

func TestForcePublic(t *testing.T) {
	c := somestruct{a: 42}
	v := reflect.ValueOf(c)
	// Without the call to forceExport(), the following line would crash with
	// "panic: reflect.Value.Interface: cannot return value obtained from
	// unexported field or method".
	a := ForceExport(v.FieldByName("a")).Interface()
	if i, ok := a.(uint32); !ok {
		t.Fatalf("Should have gotten a uint32 but got a %T", a)
	} else if i != 42 {
		t.Fatalf("Should have gotten 42 but got a %d", i)
	}
}
