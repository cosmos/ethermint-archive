// Copyright (C) 2016  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package elasticsearch

import (
	"testing"
)

func TestEscapeFieldName(t *testing.T) {
	testcases := []struct {
		unescaped string
		escaped   string
	}{{
		unescaped: "239.255.255.250_32_0.0.0.0_0",
		escaped:   "239_255_255_250_32_0_0_0_0_0",
	}, {
		unescaped: "foo",
		escaped:   "foo",
	},
	}
	for _, test := range testcases {
		escaped := EscapeFieldName(test.unescaped)
		if escaped != test.escaped {
			t.Errorf("Failed to escape %q: expected %q, got %q", test.unescaped,
				test.escaped, escaped)
		}
	}
}
