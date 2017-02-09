// Copyright (C) 2015  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package test

import (
	"testing"

	"github.com/aristanetworks/goarista/key"
)

type builtinCompare struct {
	a uint32
	b string
}

type complexCompare struct {
	m map[builtinCompare]int8
	p *complexCompare
}

type partialCompare struct {
	a uint32
	b string `deepequal:"ignore"`
}

type deepEqualTestCase struct {
	a, b interface{}
	diff string
}

type code int32

type message string

func getDeepEqualTests(t *testing.T) []deepEqualTestCase {
	var deepEqualNullMapString map[string]interface{}
	recursive := &complexCompare{}
	recursive.p = recursive
	return []deepEqualTestCase{{
		a: nil,
		b: nil,
	}, {
		a: uint8(5),
		b: uint8(5),
	}, {
		a:    nil,
		b:    uint8(5),
		diff: "expected nil but got a uint8: 0x5",
	}, {
		a:    uint8(5),
		b:    nil,
		diff: "expected a uint8 (0x5) but got nil",
	}, {
		a:    uint16(1),
		b:    uint16(2),
		diff: "uint16(1) != uint16(2)",
	}, {
		a:    int8(1),
		b:    int16(1),
		diff: "expected a int8 but got a int16",
	}, {
		a: true,
		b: true,
	}, {
		a: float32(3.1415),
		b: float32(3.1415),
	}, {
		a:    float32(3.1415),
		b:    float32(3.1416),
		diff: "float32(3.1415) != float32(3.1416)",
	}, {
		a: float64(3.14159265),
		b: float64(3.14159265),
	}, {
		a:    float64(3.14159265),
		b:    float64(3.14159266),
		diff: "float64(3.14159265) != float64(3.14159266)",
	}, {
		a: deepEqualNullMapString,
		b: deepEqualNullMapString,
	}, {
		a: &deepEqualNullMapString,
		b: &deepEqualNullMapString,
	}, {
		a:    deepEqualNullMapString,
		b:    &deepEqualNullMapString,
		diff: "expected a map[string]interface {} but got a *map[string]interface {}",
	}, {
		a:    &deepEqualNullMapString,
		b:    deepEqualNullMapString,
		diff: "expected a *map[string]interface {} but got a map[string]interface {}",
	}, {
		a: map[string]interface{}{"a": uint32(42)},
		b: map[string]interface{}{"a": uint32(42)},
	}, {
		a:    map[string]interface{}{"a": int32(42)},
		b:    map[string]interface{}{"a": int32(51)},
		diff: `for key "a" in map, values are different: int32(42) != int32(51)`,
	}, {
		a:    map[string]interface{}{"a": uint32(42)},
		b:    map[string]interface{}{},
		diff: `Maps have different size: 1 != 0 (missing key: "a")`,
	}, {
		a:    map[string]interface{}{},
		b:    map[string]interface{}{"a": uint32(42)},
		diff: `Maps have different size: 0 != 1 (extra key: "a")`,
	}, {
		a:    map[string]interface{}{"a": uint64(42), "b": "extra"},
		b:    map[string]interface{}{"a": uint64(42)},
		diff: `Maps have different size: 2 != 1 (missing key: "b")`,
	}, {
		a:    map[string]interface{}{"a": uint64(42)},
		b:    map[string]interface{}{"a": uint64(42), "b": "extra"},
		diff: `Maps have different size: 1 != 2 (extra key: "b")`,
	}, {
		a: map[uint64]interface{}{uint64(42): "foo"},
		b: map[uint64]interface{}{uint64(42): "foo"},
	}, {
		a:    map[uint64]interface{}{uint64(42): "foo"},
		b:    map[uint64]interface{}{uint64(51): "foo"},
		diff: "key uint64(42) in map is missing in the actual map",
	}, {
		a:    map[uint64]interface{}{uint64(42): "foo"},
		b:    map[uint64]interface{}{uint64(42): "foo", uint64(51): "bar"},
		diff: `Maps have different size: 1 != 2 (extra key: uint64(51))`,
	}, {
		a:    map[uint64]interface{}{uint64(42): "foo"},
		b:    map[interface{}]interface{}{uint32(42): "foo"},
		diff: "expected a map[uint64]interface {} but got a map[interface {}]interface {}",
	}, {
		a:    map[interface{}]interface{}{"a": uint32(42)},
		b:    map[string]interface{}{"a": uint32(42)},
		diff: "expected a map[interface {}]interface {} but got a map[string]interface {}",
	}, {
		a: map[interface{}]interface{}{},
		b: map[interface{}]interface{}{},
	}, {
		a: &map[interface{}]interface{}{},
		b: &map[interface{}]interface{}{},
	}, {
		a: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": int16(8)}: "foo"},
		b: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": int16(8)}: "foo"},
	}, {
		a: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": uint32(8)}: "foo"},
		b: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": uint32(8)}: "fox"},
		diff: `for complex key *map[string]interface {}{"a":"foo", "b":uint32(8)}` +
			` in map, values are different: string(foo) != string(fox)`,
	}, {
		a: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": uint32(8)}: "foo"},
		b: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": uint32(5)}: "foo"},
		diff: `complex key *map[string]interface {}{"a":"foo", "b":uint32(8)}` +
			` in map is missing in the actual map`,
	}, {
		a: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": uint32(8)}: "foo"},
		b: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo"}: "foo"},
		diff: `complex key *map[string]interface {}{"a":"foo", "b":uint32(8)}` +
			` in map is missing in the actual map`,
	}, {
		a: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": int16(8)}: "foo",
			&map[string]interface{}{"a": "foo", "b": int8(8)}:  "foo",
		},
		b: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": int16(8)}: "foo",
			&map[string]interface{}{"a": "foo", "b": int8(8)}:  "foo",
		},
	}, {
		a: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": int16(8)}: "foo",
			&map[string]interface{}{"a": "foo", "b": int8(8)}:  "foo",
		},
		b: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": int16(8)}: "foo",
			&map[string]interface{}{"a": "foo", "b": int8(5)}:  "foo",
		},
		diff: `complex key *map[string]interface {}{"a":"foo", "b":int8(8)}` +
			` in map is missing in the actual map`,
	}, {
		a: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": int16(8)}: "foo",
			&map[string]interface{}{"a": "foo", "b": int8(8)}:  "foo",
		},
		b: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": int16(8)}: "foo",
			&map[string]interface{}{"a": "foo", "b": int32(8)}: "foo",
		},
		diff: `complex key *map[string]interface {}{"a":"foo", "b":int8(8)}` +
			` in map is missing in the actual map`,
	}, {
		a: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": int16(8)}: "foo",
			&map[string]interface{}{"a": "foo", "b": int8(8)}:  "foo",
		},
		b: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": int16(8)}: "foo",
		},
		diff: `Maps have different size: 2 != 1` +
			` (extra key: *map[string]interface {}{"a":"foo", "b":int16(8)},` +
			` missing key: *map[string]interface {}{"a":"foo", "b":int16(8)},` +
			` missing key: *map[string]interface {}{"a":"foo", "b":int8(8)})`,
	}, {
		a: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": int16(8)}: "foo",
		},
		b: map[interface{}]interface{}{
			&map[string]interface{}{"a": "foo", "b": int16(8)}: "foo",
			&map[string]interface{}{"a": "foo", "b": int8(8)}:  "foo",
		},
		diff: `Maps have different size: 1 != 2` +
			` (extra key: *map[string]interface {}{"a":"foo", "b":int16(8)},` +
			` extra key: *map[string]interface {}{"a":"foo", "b":int8(8)},` +
			` missing key: *map[string]interface {}{"a":"foo", "b":int16(8)})`,
	}, {
		a: []string{},
		b: []string{},
	}, {
		a: []string{"foo", "bar"},
		b: []string{"foo", "bar"},
	}, {
		a:    []string{"foo", "bar"},
		b:    []string{"foo"},
		diff: "Expected an array of size 2 but got 1",
	}, {
		a:    []string{"foo"},
		b:    []string{"foo", "bar"},
		diff: "Expected an array of size 1 but got 2",
	}, {
		a: []string{"foo", "bar"},
		b: []string{"bar", "foo"},
		diff: `In arrays, values are different at index 0:` +
			` string(foo) != string(bar)`,
	}, {
		a:    &[]string{},
		b:    []string{},
		diff: "expected a *[]string but got a []string",
	}, {
		a: &[]string{},
		b: &[]string{},
	}, {
		a: &[]string{"foo", "bar"},
		b: &[]string{"foo", "bar"},
	}, {
		a:    &[]string{"foo", "bar"},
		b:    &[]string{"foo"},
		diff: "Expected an array of size 2 but got 1",
	}, {
		a:    &[]string{"foo"},
		b:    &[]string{"foo", "bar"},
		diff: "Expected an array of size 1 but got 2",
	}, {
		a: &[]string{"foo", "bar"},
		b: &[]string{"bar", "foo"},
		diff: `In arrays, values are different at index 0:` +
			` string(foo) != string(bar)`,
	}, {
		a: []uint32{42, 51},
		b: []uint32{42, 51},
	}, {
		a:    []uint32{42, 51},
		b:    []uint32{42, 88},
		diff: "In arrays, values are different at index 1: uint32(51) != uint32(88)",
	}, {
		a:    []uint32{42, 51},
		b:    []uint32{42},
		diff: "Expected an array of size 2 but got 1",
	}, {
		a:    []uint32{42, 51},
		b:    []uint64{42, 51},
		diff: "expected a []uint32 but got a []uint64",
	}, {
		a:    []uint64{42, 51},
		b:    []uint32{42, 51},
		diff: "expected a []uint64 but got a []uint32",
	}, {
		a: []uint64{42, 51},
		b: []uint64{42, 51},
	}, {
		a:    []uint64{42, 51},
		b:    []uint64{42},
		diff: "Expected an array of size 2 but got 1",
	}, {
		a:    []uint64{42, 51},
		b:    []uint64{42, 88},
		diff: "In arrays, values are different at index 1: uint64(51) != uint64(88)",
	}, {
		a: []interface{}{"foo", uint32(42)},
		b: []interface{}{"foo", uint32(42)},
	}, {
		a:    []interface{}{"foo", uint32(42)},
		b:    []interface{}{"foo"},
		diff: "Expected an array of size 2 but got 1",
	}, {
		a:    []interface{}{"foo"},
		b:    []interface{}{"foo", uint32(42)},
		diff: "Expected an array of size 1 but got 2",
	}, {
		a: []interface{}{"foo", uint32(42)},
		b: []interface{}{"foo", uint8(42)},
		diff: "In arrays, values are different at index 1:" +
			" expected a uint32 but got a uint8",
	}, {
		a:    []interface{}{"foo", "bar"},
		b:    []string{"foo", "bar"},
		diff: "expected a []interface {} but got a []string",
	}, {
		a: &[]interface{}{"foo", uint32(42)},
		b: &[]interface{}{"foo", uint32(42)},
	}, {
		a:    &[]interface{}{"foo", uint32(42)},
		b:    []interface{}{"foo", uint32(42)},
		diff: "expected a *[]interface {} but got a []interface {}",
	}, {
		a: comparableStruct{a: 42},
		b: comparableStruct{a: 42},
	}, {
		a: comparableStruct{a: 42, t: t},
		b: comparableStruct{a: 42},
	}, {
		a: comparableStruct{a: 42},
		b: comparableStruct{a: 42, t: t},
	}, {
		a: comparableStruct{a: 42},
		b: comparableStruct{a: 51},
		diff: "Comparable types are different: test.comparableStruct{a:" +
			"uint32(42), t:*nil} vs test.comparableStruct{a:uint32(51), t:*nil}",
	}, {
		a: builtinCompare{a: 42, b: "foo"},
		b: builtinCompare{a: 42, b: "foo"},
	}, {
		a:    builtinCompare{a: 42, b: "foo"},
		b:    builtinCompare{a: 42, b: "bar"},
		diff: `attributes "b" are different: string(foo) != string(bar)`,
	}, {
		a: map[int8]int8{2: 3, 3: 4},
		b: map[int8]int8{2: 3, 3: 4},
	}, {
		a:    map[int8]int8{2: 3, 3: 4},
		b:    map[int8]int8{2: 3, 3: 5},
		diff: "for key int8(3) in map, values are different: int8(4) != int8(5)",
	}, {
		a: complexCompare{},
		b: complexCompare{},
	}, {
		a: complexCompare{
			m: map[builtinCompare]int8{builtinCompare{1, "foo"}: 42}},
		b: complexCompare{
			m: map[builtinCompare]int8{builtinCompare{1, "foo"}: 42}},
	}, {
		a: complexCompare{
			m: map[builtinCompare]int8{builtinCompare{1, "foo"}: 42}},
		b: complexCompare{
			m: map[builtinCompare]int8{builtinCompare{1, "foo"}: 51}},
		diff: `attributes "m" are different: for key test.builtinCompare{a:uint32(1),` +
			` b:"foo"} in map, values are different: int8(42) != int8(51)`,
	}, {
		a: complexCompare{
			m: map[builtinCompare]int8{builtinCompare{1, "foo"}: 42}},
		b: complexCompare{
			m: map[builtinCompare]int8{builtinCompare{1, "bar"}: 42}},
		diff: `attributes "m" are different: key test.builtinCompare{a:uint32(1),` +
			` b:"foo"} in map is missing in the actual map`,
	}, {
		a: recursive,
		b: recursive,
	}, {
		a: complexCompare{p: recursive},
		b: complexCompare{p: recursive},
	}, {
		a: complexCompare{p: &complexCompare{p: recursive}},
		b: complexCompare{p: &complexCompare{p: recursive}},
	}, {
		a: []complexCompare{{p: &complexCompare{p: recursive}}},
		b: []complexCompare{{p: &complexCompare{p: recursive}}},
	}, {
		a: []complexCompare{{p: &complexCompare{p: recursive}}},
		b: []complexCompare{{p: &complexCompare{p: nil}}},
		diff: `In arrays, values are different at index 0: attributes "p" are` +
			` different: attributes "p" are different: got nil instead of ` +
			`*test.complexCompare{m:map[test.builtinCompare]int8{},` +
			` p:*test.complexCompare{<circular dependency>}}`,
	}, {
		a: partialCompare{a: 42},
		b: partialCompare{a: 42},
	}, {
		a:    partialCompare{a: 42},
		b:    partialCompare{a: 51},
		diff: `attributes "a" are different: uint32(42) != uint32(51)`,
	}, {
		a: partialCompare{a: 42, b: "foo"},
		b: partialCompare{a: 42, b: "bar"},
	}, {
		a: map[*builtinCompare]uint32{&builtinCompare{1, "foo"}: 42},
		b: map[*builtinCompare]uint32{&builtinCompare{1, "foo"}: 42},
	}, {
		a: map[*builtinCompare]uint32{&builtinCompare{1, "foo"}: 42},
		b: map[*builtinCompare]uint32{&builtinCompare{2, "foo"}: 42},
		diff: `complex key *test.builtinCompare{a:uint32(1), b:"foo"}` +
			` in map is missing in the actual map`,
	}, {
		a: map[*builtinCompare]uint32{&builtinCompare{1, "foo"}: 42},
		b: map[*builtinCompare]uint32{&builtinCompare{1, "foo"}: 51},
		diff: `for complex key *test.builtinCompare{a:uint32(1), b:"foo"}` +
			` in map, values are different: uint32(42) != uint32(51)`,
	}, {
		a: key.New("a"),
		b: key.New("a"),
	}, {
		a: map[key.Key]string{key.New("a"): "b"},
		b: map[key.Key]string{key.New("a"): "b"},
	}, {
		a: map[key.Key]string{key.New(map[string]interface{}{"a": true}): "b"},
		b: map[key.Key]string{key.New(map[string]interface{}{"a": true}): "b"},
	}, {
		a: key.New(map[string]interface{}{
			"a": map[key.Key]interface{}{key.New(map[string]interface{}{"k": 42}): true}}),
		b: key.New(map[string]interface{}{
			"a": map[key.Key]interface{}{key.New(map[string]interface{}{"k": 42}): true}}),
	}, {
		a: key.New(map[string]interface{}{
			"a": map[key.Key]interface{}{key.New(map[string]interface{}{"k": 42}): true}}),
		b: key.New(map[string]interface{}{
			"a": map[key.Key]interface{}{key.New(map[string]interface{}{"k": 51}): true}}),
		diff: `Comparable types are different: ` +
			`key.composite{sentinel:uintptr(18379810577513696751), m:map[string]interface {}` +
			`{"a":map[key.Key]interface {}{<max_depth>:<max_depth>}}} vs` +
			` key.composite{sentinel:uintptr(18379810577513696751), m:map[string]interface {}` +
			`{"a":map[key.Key]interface {}{<max_depth>:<max_depth>}}}`,
	}, {
		a: code(42),
		b: code(42),
	}, {
		a:    code(42),
		b:    code(51),
		diff: "code(42) != code(51)",
	}, {
		a: message("foo"),
		b: message("foo"),
	}, {
		a:    message("foo"),
		b:    message("bar"),
		diff: `message("foo") != message("bar")`,
	}, {
		a: []byte("foo"),
		b: []byte("foo"),
	}, {
		a:    []byte("foo"),
		b:    []byte("bar"),
		diff: `[]byte("foo") != []byte("bar")`,
	}}
}
