/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"testing"
)

func TestCalculateActivePumps(t *testing.T) {
	cases := []struct {
		pressure float64
		minPumpCount int64
		maxPumpCount int64
	} {
		{ pressure: 11.0, minPumpCount: 0, maxPumpCount: 2},
		{ pressure: 9.0, minPumpCount: 4, maxPumpCount: 6},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("pressure:%f", c.pressure), func(t *testing.T) {
			pumps := calculateActivePumps(c.pressure)
			if pumps == nil {
				t.Fatalf("Expected calculateActivePumps() to be implemented and return a non-nil value")
			}
			if *pumps < c.minPumpCount || *pumps > c.maxPumpCount {
				t.Fatalf("Expected calculateActivePumps() for pressure %f to return a value between %d and %d, but got %d", c.pressure, c.minPumpCount, c.maxPumpCount, *pumps)
			}
		})
	}
}