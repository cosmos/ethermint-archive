// Copyright (C) 2016  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package netns

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type mockHandle int

func (mh mockHandle) close() error {
	return nil
}

func (mh mockHandle) fd() int {
	return 0
}

func TestNetNs(t *testing.T) {
	setNsCallCount := 0

	// Mock getNs
	oldGetNs := getNs
	getNs = func(nsName string) (handle, error) {
		return mockHandle(1), nil
	}
	defer func() {
		getNs = oldGetNs
	}()

	// Mock setNs
	oldSetNs := setNs
	setNs = func(fd handle) error {
		setNsCallCount++
		return nil
	}
	defer func() {
		setNs = oldSetNs
	}()

	// Create a tempfile so we can use its name for the network namespace
	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("Failed to create a temp file: %s", err)
	}
	defer os.Remove(tmpfile.Name())
	nsName := filepath.Base(tmpfile.Name())

	// Map of network namespace name to the number of times it should call setNs
	cases := map[string]int{"": 0, "default": 2, nsName: 2}
	for name, callCount := range cases {
		var cbResult string
		err = Do(name, func() error {
			cbResult = "Hello" + name
			return nil
		})
		if err != nil {
			t.Fatalf("Error calling function in different network namespace: %s", err)
		}
		if cbResult != "Hello"+name {
			t.Fatalf("Failed to call the callback function")
		}
		if setNsCallCount != callCount {
			t.Fatalf("setNs should have been called %d times for %s, but was called %d times",
				callCount, name, setNsCallCount)
		}
		setNsCallCount = 0
	}
}
