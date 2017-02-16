// Copyright (C) 2016  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package openconfig

import (
	"encoding/json"
	"testing"

	"github.com/aristanetworks/goarista/test"
	"github.com/openconfig/reference/rpc/openconfig"
)

func TestNotificationToMap(t *testing.T) {
	value := map[string]interface{}{
		"239.255.255.250_0.0.0.0": map[string]interface{}{
			"creationTime": 4.567969230573434e+06,
		},
	}
	valueJSON, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		notification openconfig.Notification
		json         map[string]interface{}
	}{{
		notification: openconfig.Notification{
			Prefix: &openconfig.Path{
				Element: []string{
					"foo",
				},
			},
			Update: []*openconfig.Update{
				{
					Path: &openconfig.Path{
						Element: []string{
							"route1",
						},
					},
					Value: &openconfig.Value{
						Value: valueJSON,
					},
				}, {
					Path: &openconfig.Path{
						Element: []string{
							"route2",
						},
					},
					Value: &openconfig.Value{
						Value: valueJSON,
					},
				}},
		},
		json: map[string]interface{}{
			"timestamp": int64(0),
			"dataset":   "cairo",
			"update": map[string]interface{}{
				"foo": map[string]interface{}{
					"route1": map[string]interface{}{
						"239.255.255.250_0.0.0.0": map[string]interface{}{
							"creationTime": 4.567969230573434e+06,
						},
					},
					"route2": map[string]interface{}{
						"239.255.255.250_0.0.0.0": map[string]interface{}{
							"creationTime": 4.567969230573434e+06,
						},
					},
				},
			},
		},
	}, {
		notification: openconfig.Notification{
			Prefix: &openconfig.Path{
				Element: []string{
					"foo", "bar",
				},
			},
			Delete: []*openconfig.Path{
				&openconfig.Path{
					Element: []string{
						"route", "237.255.255.250_0.0.0.0",
					}},
				&openconfig.Path{
					Element: []string{
						"route", "238.255.255.250_0.0.0.0",
					},
				},
			},
			Update: []*openconfig.Update{{
				Path: &openconfig.Path{
					Element: []string{
						"route",
					},
				},
				Value: &openconfig.Value{
					Value: valueJSON,
				},
			}},
		},
		json: map[string]interface{}{
			"timestamp": int64(0),
			"dataset":   "cairo",
			"delete": map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"route": map[string]interface{}{
							"237.255.255.250_0.0.0.0": map[string]interface{}{},
							"238.255.255.250_0.0.0.0": map[string]interface{}{},
						},
					},
				},
			},
			"update": map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"route": map[string]interface{}{
							"239.255.255.250_0.0.0.0": map[string]interface{}{
								"creationTime": 4.567969230573434e+06,
							},
						},
					},
				},
			},
		},
	}}
	for _, tcase := range tests {
		actual, err := NotificationToMap("cairo", &tcase.notification, nil)
		if err != nil {
			t.Fatal(err)
		}
		diff := test.Diff(tcase.json, actual)
		if len(diff) > 0 {
			expectedJSON, _ := json.Marshal(tcase.json)
			actualJSON, _ := json.Marshal(actual)
			t.Fatalf("Unexpected diff: %s\nExpected:\n%s\nGot:\n%s\n)", diff, expectedJSON,
				actualJSON)
		}
	}
}
