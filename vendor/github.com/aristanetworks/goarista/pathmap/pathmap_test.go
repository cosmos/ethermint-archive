// Copyright (C) 2016  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package pathmap

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aristanetworks/goarista/test"
)

func accumulator(counter map[int]int) VisitorFunc {
	return func(val interface{}) error {
		counter[val.(int)]++
		return nil
	}
}

func TestVisit(t *testing.T) {
	m := New()
	m.Set([]string{"foo", "bar", "baz"}, 1)
	m.Set([]string{"*", "bar", "baz"}, 2)
	m.Set([]string{"*", "*", "baz"}, 3)
	m.Set([]string{"*", "*", "*"}, 4)
	m.Set([]string{"foo", "*", "*"}, 5)
	m.Set([]string{"foo", "bar", "*"}, 6)
	m.Set([]string{"foo", "*", "baz"}, 7)
	m.Set([]string{"*", "bar", "*"}, 8)

	m.Set([]string{}, 10)

	m.Set([]string{"*"}, 20)
	m.Set([]string{"foo"}, 21)

	m.Set([]string{"zap", "zip"}, 30)
	m.Set([]string{"zap", "zip"}, 31)

	m.Set([]string{"zip", "*"}, 40)
	m.Set([]string{"zip", "*"}, 41)

	testCases := []struct {
		path     []string
		expected map[int]int
	}{{
		path:     []string{"foo", "bar", "baz"},
		expected: map[int]int{1: 1, 2: 1, 3: 1, 4: 1, 5: 1, 6: 1, 7: 1, 8: 1},
	}, {
		path:     []string{"qux", "bar", "baz"},
		expected: map[int]int{2: 1, 3: 1, 4: 1, 8: 1},
	}, {
		path:     []string{"foo", "qux", "baz"},
		expected: map[int]int{3: 1, 4: 1, 5: 1, 7: 1},
	}, {
		path:     []string{"foo", "bar", "qux"},
		expected: map[int]int{4: 1, 5: 1, 6: 1, 8: 1},
	}, {
		path:     []string{},
		expected: map[int]int{10: 1},
	}, {
		path:     []string{"foo"},
		expected: map[int]int{20: 1, 21: 1},
	}, {
		path:     []string{"foo", "bar"},
		expected: map[int]int{},
	}, {
		path:     []string{"zap", "zip"},
		expected: map[int]int{31: 1},
	}, {
		path:     []string{"zip", "zap"},
		expected: map[int]int{41: 1},
	}}

	for _, tc := range testCases {
		result := make(map[int]int, len(tc.expected))
		m.Visit(tc.path, accumulator(result))
		if diff := test.Diff(tc.expected, result); diff != "" {
			t.Errorf("Test case %v: %s", tc.path, diff)
		}
	}
}

func TestVisitError(t *testing.T) {
	m := New()
	m.Set([]string{"foo", "bar"}, 1)
	m.Set([]string{"*", "bar"}, 2)

	errTest := errors.New("Test")

	err := m.Visit([]string{"foo", "bar"}, func(v interface{}) error { return errTest })
	if err != errTest {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", errTest, err)
	}
	err = m.VisitPrefix([]string{"foo", "bar", "baz"}, func(v interface{}) error { return errTest })
	if err != errTest {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", errTest, err)
	}
}

func TestGet(t *testing.T) {
	m := New()
	m.Set([]string{}, 0)
	m.Set([]string{"foo", "bar"}, 1)
	m.Set([]string{"foo", "*"}, 2)
	m.Set([]string{"*", "bar"}, 3)
	m.Set([]string{"zap", "zip"}, 4)

	testCases := []struct {
		path     []string
		expected interface{}
	}{{
		path:     []string{},
		expected: 0,
	}, {
		path:     []string{"foo", "bar"},
		expected: 1,
	}, {
		path:     []string{"foo", "*"},
		expected: 2,
	}, {
		path:     []string{"*", "bar"},
		expected: 3,
	}, {
		path:     []string{"bar", "foo"},
		expected: nil,
	}, {
		path:     []string{"zap", "*"},
		expected: nil,
	}}

	for _, tc := range testCases {
		got := m.Get(tc.path)
		if got != tc.expected {
			t.Errorf("Test case %v: Expected %v, Got %v",
				tc.path, tc.expected, got)
		}
	}
}

func countNodes(n *node) int {
	if n == nil {
		return 0
	}
	count := 1
	count += countNodes(n.wildcard)
	for _, child := range n.children {
		count += countNodes(child)
	}
	return count
}

func TestDelete(t *testing.T) {
	m := New()
	m.Set([]string{}, 0)
	m.Set([]string{"*"}, 1)
	m.Set([]string{"foo", "bar"}, 2)
	m.Set([]string{"foo", "*"}, 3)

	n := countNodes(m.(*node))
	if n != 5 {
		t.Errorf("Initial count wrong. Expected: 5, Got: %d", n)
	}

	testCases := []struct {
		del      []string    // Path to delete
		expected bool        // expected return value of Delete
		visit    []string    // Path to Visit
		before   map[int]int // Expected to find items before deletion
		after    map[int]int // Expected to find items after deletion
		count    int         // Count of nodes
	}{{
		del:      []string{"zap"}, // A no-op Delete
		expected: false,
		visit:    []string{"foo", "bar"},
		before:   map[int]int{2: 1, 3: 1},
		after:    map[int]int{2: 1, 3: 1},
		count:    5,
	}, {
		del:      []string{"foo", "bar"},
		expected: true,
		visit:    []string{"foo", "bar"},
		before:   map[int]int{2: 1, 3: 1},
		after:    map[int]int{3: 1},
		count:    4,
	}, {
		del:      []string{"*"},
		expected: true,
		visit:    []string{"foo"},
		before:   map[int]int{1: 1},
		after:    map[int]int{},
		count:    3,
	}, {
		del:      []string{"*"},
		expected: false,
		visit:    []string{"foo"},
		before:   map[int]int{},
		after:    map[int]int{},
		count:    3,
	}, {
		del:      []string{"foo", "*"},
		expected: true,
		visit:    []string{"foo", "bar"},
		before:   map[int]int{3: 1},
		after:    map[int]int{},
		count:    1, // Should have deleted "foo" and "bar" nodes
	}, {
		del:      []string{},
		expected: true,
		visit:    []string{},
		before:   map[int]int{0: 1},
		after:    map[int]int{},
		count:    1, // Root node can't be deleted
	}}

	for i, tc := range testCases {
		beforeResult := make(map[int]int, len(tc.before))
		m.Visit(tc.visit, accumulator(beforeResult))
		if diff := test.Diff(tc.before, beforeResult); diff != "" {
			t.Errorf("Test case %d (%v): %s", i, tc.del, diff)
		}

		if got := m.Delete(tc.del); got != tc.expected {
			t.Errorf("Test case %d (%v): Unexpected return. Expected %t, Got: %t",
				i, tc.del, tc.expected, got)
		}

		afterResult := make(map[int]int, len(tc.after))
		m.Visit(tc.visit, accumulator(afterResult))
		if diff := test.Diff(tc.after, afterResult); diff != "" {
			t.Errorf("Test case %d (%v): %s", i, tc.del, diff)
		}
	}
}

func TestVisitPrefix(t *testing.T) {
	m := New()
	m.Set([]string{}, 0)
	m.Set([]string{"foo"}, 1)
	m.Set([]string{"foo", "bar"}, 2)
	m.Set([]string{"foo", "bar", "baz"}, 3)
	m.Set([]string{"foo", "bar", "baz", "quux"}, 4)
	m.Set([]string{"quux", "bar"}, 5)
	m.Set([]string{"foo", "quux"}, 6)
	m.Set([]string{"*"}, 7)
	m.Set([]string{"foo", "*"}, 8)
	m.Set([]string{"*", "bar"}, 9)
	m.Set([]string{"*", "quux"}, 10)
	m.Set([]string{"quux", "quux", "quux", "quux"}, 11)

	testCases := []struct {
		path     []string
		expected map[int]int
	}{{
		path:     []string{"foo", "bar", "baz"},
		expected: map[int]int{0: 1, 1: 1, 2: 1, 3: 1, 7: 1, 8: 1, 9: 1},
	}, {
		path:     []string{"zip", "zap"},
		expected: map[int]int{0: 1, 7: 1},
	}, {
		path:     []string{"foo", "zap"},
		expected: map[int]int{0: 1, 1: 1, 8: 1, 7: 1},
	}, {
		path:     []string{"quux", "quux", "quux"},
		expected: map[int]int{0: 1, 7: 1, 10: 1},
	}}

	for _, tc := range testCases {
		result := make(map[int]int, len(tc.expected))
		m.VisitPrefix(tc.path, accumulator(result))
		if diff := test.Diff(tc.expected, result); diff != "" {
			t.Errorf("Test case %v: %s", tc.path, diff)
		}
	}
}

func TestString(t *testing.T) {
	m := New()
	m.Set([]string{}, 0)
	m.Set([]string{"foo", "bar"}, 1)
	m.Set([]string{"foo", "quux"}, 2)
	m.Set([]string{"foo", "*"}, 3)

	expected := `Val: 0
Child "foo":
  Child "*":
    Val: 3
  Child "bar":
    Val: 1
  Child "quux":
    Val: 2
`
	got := fmt.Sprint(m)

	if expected != got {
		t.Errorf("Unexpected string. Expected:\n\n%s\n\nGot:\n\n%s", expected, got)
	}
}

func genWords(count, wordLength int) []string {
	chars := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	if count+wordLength > len(chars) {
		panic("need more chars")
	}
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = string(chars[i : i+wordLength])
	}
	return result
}

func benchmarkPathMap(pathLength, pathDepth int, b *testing.B) {
	m := New()

	// Push pathDepth paths, each of length pathLength
	path := genWords(pathLength, 10)
	words := genWords(pathDepth, 10)
	n := m.(*node)
	for _, element := range path {
		n.children = map[string]*node{}
		for _, word := range words {
			n.children[word] = &node{}
		}
		n = n.children[element]
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Visit(path, func(v interface{}) error { return nil })
	}
}

func BenchmarkPathMap1x25(b *testing.B)  { benchmarkPathMap(1, 25, b) }
func BenchmarkPathMap10x50(b *testing.B) { benchmarkPathMap(10, 25, b) }
func BenchmarkPathMap20x50(b *testing.B) { benchmarkPathMap(20, 25, b) }
