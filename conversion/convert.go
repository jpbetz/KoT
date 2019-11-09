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
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog"

	v1 "github.com/jpbetz/KoT/apis/things/v1"
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
)

func convertValueToV1alpha(in *v1.Value) *v1alpha1.Value {
	if in == nil {
		return nil
	}
	ret := &v1alpha1.Value{
		Name: in.Name,
	}

	switch {
	case in.Boolean != nil:
		ret.Type = v1alpha1.BooleanType
		if *in.Boolean {
			ret.Value = *resource.NewQuantity(1, resource.DecimalSI)
		} else {
			ret.Value = *resource.NewQuantity(0, resource.DecimalSI)
		}
	case in.Float != nil:
		ret.Type = v1alpha1.FloatType
		ret.Value = *in.Float
	case in.Integer != nil:
		ret.Type = v1alpha1.IntegerType
		ret.Value = *resource.NewQuantity(int64(*in.Integer), resource.DecimalSI)
	}

	return ret
}

func convertValueToV1(in *v1alpha1.Value) *v1.Value {
	if in == nil {
		return nil
	}

	ret := &v1.Value{
		Name: in.Name,
	}
	switch in.Type {
	case v1alpha1.BooleanType:
		b := !in.Value.IsZero()
		ret.Boolean = &b
	case v1alpha1.IntegerType:
		i64, _ := in.Value.AsInt64()
		i32 := int32(i64)
		ret.Integer = &i32
	case v1alpha1.FloatType:
		clone := in.Value.DeepCopy()
		ret.Float = &clone
	}

	return ret
}

func convertValuesToV1alpha1(in []v1.Value) []v1alpha1.Value {
	if in == nil {
		return nil
	}
	ret := make([]v1alpha1.Value, 0, len(in))
	for _, v := range in {
		ret = append(ret, *convertValueToV1alpha(&v))
	}
	return ret
}

func convertValuesToV1(in []v1alpha1.Value) []v1.Value {
	if in == nil {
		return nil
	}
	ret := make([]v1.Value, 0, len(in))
	for _, v := range in {
		ret = append(ret, *convertValueToV1(&v))
	}
	return ret
}

func convert(in runtime.Object, apiVersion string) (runtime.Object, error) {
	switch in := in.(type) {
	case *v1alpha1.Device:
		if apiVersion != v1.SchemeGroupVersion.String() {
			return nil, fmt.Errorf("cannot convert %s to %s", v1alpha1.SchemeGroupVersion, apiVersion)
		}
		klog.V(2).Infof("Converting %s/%s from %s to %s", in.Namespace, in.Name, v1alpha1.SchemeGroupVersion, apiVersion)

		out := &v1.Device{
			TypeMeta:   in.TypeMeta,
			ObjectMeta: in.ObjectMeta,
			Spec:       v1.DeviceSpec{Inputs: convertValuesToV1(in.Spec.Inputs)},
			Status: v1.DeviceStatus{
				ObservedInputs: convertValuesToV1(in.Status.ObservedInputs),
				Outputs:        convertValuesToV1(in.Status.Outputs),
			},
		}
		out.TypeMeta.APIVersion = apiVersion

		if klog.V(6) {
			klog.Infof("In: %s\nOut: %s", marshal(in), marshal(out))

		}

		return out, nil

	case *v1.Device:
		if apiVersion != v1alpha1.SchemeGroupVersion.String() {
			return nil, fmt.Errorf("cannot convert %s to %s", v1.SchemeGroupVersion, apiVersion)
		}
		klog.V(2).Infof("Converting %s/%s from %s to %s", in.Namespace, in.Name, v1alpha1.SchemeGroupVersion, apiVersion)

		out := &v1alpha1.Device{
			TypeMeta:   in.TypeMeta,
			ObjectMeta: in.ObjectMeta,
			Spec:       v1alpha1.DeviceSpec{Inputs: convertValuesToV1alpha1(in.Spec.Inputs)},
			Status: v1alpha1.DeviceStatus{
				ObservedInputs: convertValuesToV1alpha1(in.Status.ObservedInputs),
				Outputs:        convertValuesToV1alpha1(in.Status.Outputs),
			},
		}
		out.TypeMeta.APIVersion = apiVersion

		if klog.V(6) {
			klog.Infof("In: %s\nOut: %s", marshal(in), marshal(out))
		}

		return out, nil

	default:
	}
	klog.V(2).Infof("Unknown type %T", in)
	return nil, fmt.Errorf("unknown type %T", in)
}

func marshal(x interface{}) string {
	ret, _ := json.Marshal(x)
	return string(ret)
}
