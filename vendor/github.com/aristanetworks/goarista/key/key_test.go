// Copyright (C) 2015  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package key_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	. "github.com/aristanetworks/goarista/key"
	"github.com/aristanetworks/goarista/test"
	"github.com/aristanetworks/goarista/value"
)

type compareMe struct {
	i int
}

func (c compareMe) Equal(other interface{}) bool {
	o, ok := other.(compareMe)
	return ok && c == o
}

type customKey struct {
	i int
}

var _ value.Value = customKey{}

func (c customKey) String() string {
	return fmt.Sprintf("customKey=%d", c.i)
}

func (c customKey) MarshalJSON() ([]byte, error) {
	return nil, nil
}

func (c customKey) ToBuiltin() interface{} {
	return c.i
}

func TestKeyEqual(t *testing.T) {
	tests := []struct {
		a      Key
		b      Key
		result bool
	}{{
		a:      New("foo"),
		b:      New("foo"),
		result: true,
	}, {
		a:      New("foo"),
		b:      New("bar"),
		result: false,
	}, {
		a:      New(map[string]interface{}{}),
		b:      New("bar"),
		result: false,
	}, {
		a:      New(map[string]interface{}{}),
		b:      New(map[string]interface{}{}),
		result: true,
	}, {
		a:      New(map[string]interface{}{"a": 3}),
		b:      New(map[string]interface{}{}),
		result: false,
	}, {
		a:      New(map[string]interface{}{"a": 3}),
		b:      New(map[string]interface{}{"b": 4}),
		result: false,
	}, {
		a:      New(map[string]interface{}{"a": 4, "b": 5}),
		b:      New(map[string]interface{}{"a": 4}),
		result: false,
	}, {
		a:      New(map[string]interface{}{"a": 3}),
		b:      New(map[string]interface{}{"a": 4}),
		result: false,
	}, {
		a:      New(map[string]interface{}{"a": 3}),
		b:      New(map[string]interface{}{"a": 3}),
		result: true,
	}, {
		a:      New(map[string]interface{}{"a": map[Key]interface{}{New("b"): 3}}),
		b:      New(map[string]interface{}{"a": map[Key]interface{}{New("b"): 4}}),
		result: false,
	}, {
		a:      New(map[string]interface{}{"a": map[Key]interface{}{New("b"): 4, New("c"): 5}}),
		b:      New(map[string]interface{}{"a": map[Key]interface{}{New("b"): 4}}),
		result: false,
	}, {
		a:      New(map[string]interface{}{"a": map[Key]interface{}{New("b"): 4, New("c"): 5}}),
		b:      New(map[string]interface{}{"a": map[Key]interface{}{New("b"): 4, New("c"): 5}}),
		result: true,
	}, {
		a:      New(map[string]interface{}{"a": map[Key]interface{}{New("b"): 4}}),
		b:      New(map[string]interface{}{"a": map[Key]interface{}{New("b"): 4}}),
		result: true,
	}, {
		a:      New(map[string]interface{}{"a": compareMe{i: 3}}),
		b:      New(map[string]interface{}{"a": compareMe{i: 3}}),
		result: true,
	}, {
		a:      New(map[string]interface{}{"a": compareMe{i: 3}}),
		b:      New(map[string]interface{}{"a": compareMe{i: 4}}),
		result: false,
	}, {
		a:      New(customKey{i: 42}),
		b:      New(customKey{i: 42}),
		result: true,
	}}

	for _, tcase := range tests {
		if tcase.a.Equal(tcase.b) != tcase.result {
			t.Errorf("Wrong result for case:\na: %#v\nb: %#v\nresult: %#v",
				tcase.a,
				tcase.b,
				tcase.result)
		}
	}

	if New("a").Equal(32) {
		t.Error("Wrong result for different types case")
	}
}

func TestGetFromMap(t *testing.T) {
	tests := []struct {
		k     Key
		m     map[Key]interface{}
		v     interface{}
		found bool
	}{{
		k:     New("a"),
		m:     map[Key]interface{}{New("a"): "b"},
		v:     "b",
		found: true,
	}, {
		k:     New(uint32(35)),
		m:     map[Key]interface{}{New(uint32(35)): "c"},
		v:     "c",
		found: true,
	}, {
		k:     New(uint32(37)),
		m:     map[Key]interface{}{New(uint32(36)): "c"},
		found: false,
	}, {
		k:     New(uint32(37)),
		m:     map[Key]interface{}{},
		found: false,
	}, {
		k: New(map[string]interface{}{"a": "b", "c": uint64(4)}),
		m: map[Key]interface{}{
			New(map[string]interface{}{"a": "b", "c": uint64(4)}): "foo",
		},
		v:     "foo",
		found: true,
	}, {
		k: New(map[string]interface{}{"a": "b", "c": uint64(4)}),
		m: map[Key]interface{}{
			New(map[string]interface{}{"a": "b", "c": uint64(5)}): "foo",
		},
		found: false,
	}, {
		k:     New(customKey{i: 42}),
		m:     map[Key]interface{}{New(customKey{i: 42}): "c"},
		v:     "c",
		found: true,
	}, {
		k:     New(customKey{i: 42}),
		m:     map[Key]interface{}{New(customKey{i: 43}): "c"},
		found: false,
	}, {
		k: New(map[string]interface{}{
			"damn": map[Key]interface{}{
				New(map[string]interface{}{"a": uint32(42),
					"b": uint32(51)}): true}}),
		m: map[Key]interface{}{
			New(map[string]interface{}{
				"damn": map[Key]interface{}{
					New(map[string]interface{}{"a": uint32(42),
						"b": uint32(51)}): true}}): "foo",
		},
		v:     "foo",
		found: true,
	}, {
		k: New(map[string]interface{}{
			"damn": map[Key]interface{}{
				New(map[string]interface{}{"a": uint32(42),
					"b": uint32(52)}): true}}),
		m: map[Key]interface{}{
			New(map[string]interface{}{
				"damn": map[Key]interface{}{
					New(map[string]interface{}{"a": uint32(42),
						"b": uint32(51)}): true}}): "foo",
		},
		found: false,
	}, {
		k: New(map[string]interface{}{
			"nested": map[string]interface{}{
				"a": uint32(42), "b": uint32(51)}}),
		m: map[Key]interface{}{
			New(map[string]interface{}{
				"nested": map[string]interface{}{
					"a": uint32(42), "b": uint32(51)}}): "foo",
		},
		v:     "foo",
		found: true,
	}, {
		k: New(map[string]interface{}{
			"nested": map[string]interface{}{
				"a": uint32(42), "b": uint32(52)}}),
		m: map[Key]interface{}{
			New(map[string]interface{}{
				"nested": map[string]interface{}{
					"a": uint32(42), "b": uint32(51)}}): "foo",
		},
		found: false,
	}}

	for _, tcase := range tests {
		v, ok := tcase.m[tcase.k]
		if tcase.found != ok {
			t.Errorf("Wrong retrieval result for case:\nk: %#v\nm: %#v\nv: %#v",
				tcase.k,
				tcase.m,
				tcase.v)
		} else if tcase.found && !ok {
			t.Errorf("Unable to retrieve value for case:\nk: %#v\nm: %#v\nv: %#v",
				tcase.k,
				tcase.m,
				tcase.v)
		} else if tcase.found && !test.DeepEqual(tcase.v, v) {
			t.Errorf("Wrong result for case:\nk: %#v\nm: %#v\nv: %#v",
				tcase.k,
				tcase.m,
				tcase.v)
		}
	}
}

func TestDeleteFromMap(t *testing.T) {
	tests := []struct {
		k Key
		m map[Key]interface{}
		r map[Key]interface{}
	}{{
		k: New("a"),
		m: map[Key]interface{}{New("a"): "b"},
		r: map[Key]interface{}{},
	}, {
		k: New("b"),
		m: map[Key]interface{}{New("a"): "b"},
		r: map[Key]interface{}{New("a"): "b"},
	}, {
		k: New("a"),
		m: map[Key]interface{}{},
		r: map[Key]interface{}{},
	}, {
		k: New(uint32(35)),
		m: map[Key]interface{}{New(uint32(35)): "c"},
		r: map[Key]interface{}{},
	}, {
		k: New(uint32(36)),
		m: map[Key]interface{}{New(uint32(35)): "c"},
		r: map[Key]interface{}{New(uint32(35)): "c"},
	}, {
		k: New(uint32(37)),
		m: map[Key]interface{}{},
		r: map[Key]interface{}{},
	}, {
		k: New(map[string]interface{}{"a": "b", "c": uint64(4)}),
		m: map[Key]interface{}{
			New(map[string]interface{}{"a": "b", "c": uint64(4)}): "foo",
		},
		r: map[Key]interface{}{},
	}, {
		k: New(customKey{i: 42}),
		m: map[Key]interface{}{New(customKey{i: 42}): "c"},
		r: map[Key]interface{}{},
	}}

	for _, tcase := range tests {
		delete(tcase.m, tcase.k)
		if !test.DeepEqual(tcase.m, tcase.r) {
			t.Errorf("Wrong result for case:\nk: %#v\nm: %#v\nr: %#v",
				tcase.k,
				tcase.m,
				tcase.r)
		}
	}
}

func TestSetToMap(t *testing.T) {
	tests := []struct {
		k Key
		v interface{}
		m map[Key]interface{}
		r map[Key]interface{}
	}{{
		k: New("a"),
		v: "c",
		m: map[Key]interface{}{New("a"): "b"},
		r: map[Key]interface{}{New("a"): "c"},
	}, {
		k: New("b"),
		v: uint64(56),
		m: map[Key]interface{}{New("a"): "b"},
		r: map[Key]interface{}{
			New("a"): "b",
			New("b"): uint64(56),
		},
	}, {
		k: New("a"),
		v: "foo",
		m: map[Key]interface{}{},
		r: map[Key]interface{}{New("a"): "foo"},
	}, {
		k: New(uint32(35)),
		v: "d",
		m: map[Key]interface{}{New(uint32(35)): "c"},
		r: map[Key]interface{}{New(uint32(35)): "d"},
	}, {
		k: New(uint32(36)),
		v: true,
		m: map[Key]interface{}{New(uint32(35)): "c"},
		r: map[Key]interface{}{
			New(uint32(35)): "c",
			New(uint32(36)): true,
		},
	}, {
		k: New(uint32(37)),
		v: false,
		m: map[Key]interface{}{New(uint32(36)): "c"},
		r: map[Key]interface{}{
			New(uint32(36)): "c",
			New(uint32(37)): false,
		},
	}, {
		k: New(uint32(37)),
		v: "foobar",
		m: map[Key]interface{}{},
		r: map[Key]interface{}{New(uint32(37)): "foobar"},
	}, {
		k: New(map[string]interface{}{"a": "b", "c": uint64(4)}),
		v: "foobar",
		m: map[Key]interface{}{
			New(map[string]interface{}{"a": "b", "c": uint64(4)}): "foo",
		},
		r: map[Key]interface{}{
			New(map[string]interface{}{"a": "b", "c": uint64(4)}): "foobar",
		},
	}, {
		k: New(map[string]interface{}{"a": "b", "c": uint64(7)}),
		v: "foobar",
		m: map[Key]interface{}{
			New(map[string]interface{}{"a": "b", "c": uint64(4)}): "foo",
		},
		r: map[Key]interface{}{
			New(map[string]interface{}{"a": "b", "c": uint64(4)}): "foo",
			New(map[string]interface{}{"a": "b", "c": uint64(7)}): "foobar",
		},
	}, {
		k: New(map[string]interface{}{"a": "b", "d": uint64(6)}),
		v: "barfoo",
		m: map[Key]interface{}{
			New(map[string]interface{}{"a": "b", "c": uint64(4)}): "foo",
		},
		r: map[Key]interface{}{
			New(map[string]interface{}{"a": "b", "c": uint64(4)}): "foo",
			New(map[string]interface{}{"a": "b", "d": uint64(6)}): "barfoo",
		},
	}, {
		k: New(customKey{i: 42}),
		v: "foo",
		m: map[Key]interface{}{},
		r: map[Key]interface{}{New(customKey{i: 42}): "foo"},
	}}

	for i, tcase := range tests {
		tcase.m[tcase.k] = tcase.v
		if !test.DeepEqual(tcase.m, tcase.r) {
			t.Errorf("Wrong result for case %d:\nk: %#v\nm: %#v\nr: %#v",
				i,
				tcase.k,
				tcase.m,
				tcase.r)
		}
	}
}

func TestMisc(t *testing.T) {
	k := New(map[string]interface{}{"foo": true})
	js, err := json.Marshal(k)
	if err != nil {
		t.Error("JSON encoding failed:", err)
	} else if expected := `{"foo":true}`; string(js) != expected {
		t.Errorf("Wanted JSON %q but got %q", expected, js)
	}
	expected := `key.New(map[string]interface {}{"foo":true})`
	gostr := fmt.Sprintf("%#v", k)
	if expected != gostr {
		t.Errorf("Wanted Go representation %q but got %q", expected, gostr)
	}

	test.ShouldPanic(t, func() { New(42) })

	k = New(customKey{i: 42})
	if expected, str := "customKey=42", k.String(); expected != str {
		t.Errorf("Wanted string representation %q but got %q", expected, str)
	}
}

func BenchmarkSetToMapWithStringKey(b *testing.B) {
	m := map[Key]interface{}{
		New("a"):   true,
		New("a1"):  true,
		New("a2"):  true,
		New("a3"):  true,
		New("a4"):  true,
		New("a5"):  true,
		New("a6"):  true,
		New("a7"):  true,
		New("a8"):  true,
		New("a9"):  true,
		New("a10"): true,
		New("a11"): true,
		New("a12"): true,
		New("a13"): true,
		New("a14"): true,
		New("a15"): true,
		New("a16"): true,
		New("a17"): true,
		New("a18"): true,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m[New(strconv.Itoa(i))] = true
	}
}

func BenchmarkSetToMapWithUint64Key(b *testing.B) {
	m := map[Key]interface{}{
		New(uint64(1)):  true,
		New(uint64(2)):  true,
		New(uint64(3)):  true,
		New(uint64(4)):  true,
		New(uint64(5)):  true,
		New(uint64(6)):  true,
		New(uint64(7)):  true,
		New(uint64(8)):  true,
		New(uint64(9)):  true,
		New(uint64(10)): true,
		New(uint64(11)): true,
		New(uint64(12)): true,
		New(uint64(13)): true,
		New(uint64(14)): true,
		New(uint64(15)): true,
		New(uint64(16)): true,
		New(uint64(17)): true,
		New(uint64(18)): true,
		New(uint64(19)): true,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m[New(uint64(i))] = true
	}
}

func BenchmarkGetFromMapWithMapKey(b *testing.B) {
	m := map[Key]interface{}{
		New(map[string]interface{}{"a": true}): true,
		New(map[string]interface{}{"b": true}): true,
		New(map[string]interface{}{"c": true}): true,
		New(map[string]interface{}{"d": true}): true,
		New(map[string]interface{}{"e": true}): true,
		New(map[string]interface{}{"f": true}): true,
		New(map[string]interface{}{"g": true}): true,
		New(map[string]interface{}{"h": true}): true,
		New(map[string]interface{}{"i": true}): true,
		New(map[string]interface{}{"j": true}): true,
		New(map[string]interface{}{"k": true}): true,
		New(map[string]interface{}{"l": true}): true,
		New(map[string]interface{}{"m": true}): true,
		New(map[string]interface{}{"n": true}): true,
		New(map[string]interface{}{"o": true}): true,
		New(map[string]interface{}{"p": true}): true,
		New(map[string]interface{}{"q": true}): true,
		New(map[string]interface{}{"r": true}): true,
		New(map[string]interface{}{"s": true}): true,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := New(map[string]interface{}{string('a' + i%19): true})
		_, found := m[key]
		if !found {
			b.Fatalf("WTF: %#v", key)
		}
	}
}

func mkKey(i int) Key {
	return New(map[string]interface{}{
		"foo": map[string]interface{}{
			"aaaa1": uint32(0),
			"aaaa2": uint32(0),
			"aaaa3": uint32(i),
		},
		"bar": map[string]interface{}{
			"nested": uint32(42),
		},
	})
}

func BenchmarkBigMapWithCompositeKeys(b *testing.B) {
	const size = 10000
	m := make(map[Key]interface{}, size)
	for i := 0; i < size; i++ {
		m[mkKey(i)] = true
	}
	k := mkKey(0)
	submap := k.Key().(map[string]interface{})["foo"].(map[string]interface{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		submap["aaaa3"] = uint32(i)
		_, found := m[k]
		if found != (i < size) {
			b.Fatalf("WTF: %#v", k)
		}
	}
}
