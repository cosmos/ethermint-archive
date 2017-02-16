// Copyright (C) 2016  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package openconfig

import (
	"encoding/json"
	"testing"

	"github.com/aristanetworks/goarista/openconfig"
	"github.com/aristanetworks/goarista/test"
	pb "github.com/openconfig/reference/rpc/openconfig"
)

func TestJsonify(t *testing.T) {
	var tests = []struct {
		notification *pb.Notification
		document     map[string]interface{}
	}{{
		notification: &pb.Notification{
			Prefix: &pb.Path{Element: []string{"Sysdb", "a"}},
			Update: []*pb.Update{
				{
					Path: &pb.Path{Element: []string{"b"}},
					Value: &pb.Value{
						Value: []byte{52, 50},
						Type:  pb.Type_JSON,
					},
				},
			},
		},
		document: map[string]interface{}{
			"timestamp": int64(0),
			"dataset":   "foo",
			"update": map[string]interface{}{
				"Sysdb": map[string]interface{}{
					"a": map[string]interface{}{
						"b": 42,
					},
				},
			},
		},
	},
	}
	for _, jsonTest := range tests {
		expected, err := json.Marshal(jsonTest.document)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := openconfig.NotificationToJSONDocument("foo",
			jsonTest.notification, nil)
		if err != nil {
			t.Error(err)
		}
		diff := test.Diff(actual, expected)
		if len(diff) > 0 {
			t.Errorf("Unexpected diff: %s", diff)
		}
	}
}
