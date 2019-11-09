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

package fuzzer

import (
	fuzz "github.com/google/gofuzz"

	"k8s.io/apimachinery/pkg/api/resource"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/utils/pointer"

	v1 "github.com/jpbetz/KoT/apis/things/v1"
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
)

// Funcs returns the fuzzer functions for the things api group.
func Funcs(codecs runtimeserializer.CodecFactory) []interface{} {
	return []interface{}{
		func(v *v1.Value, c fuzz.Continue) {
			v.Name = c.RandString()
			switch c.RandUint64() % 3 {
			case 0:
				v.Integer = pointer.Int32Ptr(c.Rand.Int31())
			case 1:
				v.Boolean = pointer.BoolPtr(c.RandBool())
			case 2:
				v.Float = resource.NewMilliQuantity(c.Int63(), resource.DecimalSI)
			}
		},
		func(v *v1alpha1.Value, c fuzz.Continue) {
			v.Name = c.RandString()
			switch c.RandUint64() % 3 {
			case 0:
				v.Type = v1alpha1.IntegerType
				v.Value = *resource.NewQuantity(int64(c.Rand.Int31()), resource.DecimalSI)
			case 1:
				v.Type = v1alpha1.BooleanType
				if c.RandBool() {
					v.Value = *resource.NewQuantity(1, resource.DecimalSI)
				} else {
					v.Value = *resource.NewQuantity(0, resource.DecimalSI)
				}
			case 2:
				v.Type = v1alpha1.FloatType
				v.Value = *resource.NewMilliQuantity(c.Int63(), resource.DecimalSI)
			}
		},
	}
}
