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

package conversion

import (
	"math/rand"
	"reflect"
	"testing"

	apitestingfuzzer "k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	metafuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/diff"

	"github.com/jpbetz/KoT/apis/things"
	"github.com/jpbetz/KoT/apis/things/fuzzer"
	thingsv1 "github.com/jpbetz/KoT/apis/things/v1"
	thingsv1alpha1 "github.com/jpbetz/KoT/apis/things/v1alpha1"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		target  string
		want    string
		wantErr bool
	}{
		{
			"empty v1alpha1",
			`{"apiVersion":"things.kubecon.io/v1alpha1","kind":"Device","spec":{}}`,
			"things.kubecon.io/v1",
			`{"apiVersion":"things.kubecon.io/v1","kind":"Device","spec":{}}`,
			false,
		},
		{
			"v1alpha1 with spec",
			`{"apiVersion":"things.kubecon.io/v1alpha1","kind":"Device","spec":{"inputs":[{"name":"Switch","type":"Float","value":"1"}]}}`,
			"things.kubecon.io/v1",
			`{"apiVersion":"things.kubecon.io/v1","kind":"Device","spec":{"inputs":[{"name":"Switch","float":"1"}]}}`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, _, err := codecs.UniversalDeserializer().Decode([]byte(tt.in), nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			expected, _, err := codecs.UniversalDeserializer().Decode([]byte(tt.want), nil, nil)
			if err != nil {
				t.Fatal(err)
			}

			got, err := convert(obj, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, expected) {
				t.Errorf("got unpexpect converted object: %s", diff.StringDiff(marshal(expected), marshal(got)))
			}

			roundtripped, err := convert(got, obj.GetObjectKind().GroupVersionKind().GroupVersion().String())
			if err != nil {
				t.Errorf("unexpected roundtrip error: %v", err)
				return
			}

			if !reflect.DeepEqual(obj, roundtripped) {
				t.Errorf("got unpexpect roundtripped object: %s", diff.StringDiff(marshal(obj), marshal(roundtripped)))
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	f := apitestingfuzzer.FuzzerFor(
		apitestingfuzzer.MergeFuzzerFuncs(metafuzzer.Funcs, fuzzer.Funcs),
		rand.NewSource(rand.Int63()),
		codecs,
	)

	for _, kind := range []string{"Device"} {
		for _, version := range []string{"v1", "v1alpha1"} {
			gvk := schema.GroupVersionKind{Group: things.GroupName, Version: version, Kind: kind}
			t.Run(gvk.Group+"."+gvk.Version+"."+gvk.Kind, func(t *testing.T) {
				for i := 0; i < 1000; i++ {
					x, err := scheme.New(gvk)
					if err != nil {
						t.Fatal(err)
					}

					f.Fuzz(x)
					x.GetObjectKind().SetGroupVersionKind(gvk)

					otherVersion := thingsv1.SchemeGroupVersion
					if gvk.Version == "v1" {
						otherVersion = thingsv1alpha1.SchemeGroupVersion
					}

					other, err := convert(x, otherVersion.String())
					if err != nil {
						t.Errorf("failed to convert %#v to %s: %v", x, otherVersion, err)
						continue
					}

					back, err := convert(other, gvk.GroupVersion().String())
					if err != nil {
						t.Errorf("failed to convert %#v back to %s: %v", other, gvk.Version, err)
						continue
					}

					if jx, jback := marshal(x), marshal(back); jx != jback {
						t.Errorf("roundtrip failed: %s", diff.StringDiff(jx, jback))
					}
				}
			})
		}
	}
}
