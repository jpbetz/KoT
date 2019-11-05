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
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
)

func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

func getObservedInput(device v1alpha1.Device, inputName string) (v1alpha1.Value, bool) {
	for _, input := range device.Status.ObservedInputs {
		if input.Name == inputName {
			return input, true
		}
	}
	return v1alpha1.Value{}, false
}

func getOutput(device v1alpha1.Device, outputName string) (v1alpha1.Value, bool) {
	for _, output := range device.Status.Outputs {
		if output.Name == outputName {
			return output, true
		}
	}
	return v1alpha1.Value{}, false
}

func getInput(device v1alpha1.Device, inputName string) (v1alpha1.Value, bool) {
	for _, input := range device.Spec.Inputs {
		if input.Name == inputName {
			return input, true
		}
	}
	return v1alpha1.Value{}, false
}

func setInputValue(device v1alpha1.Device, inputName string, value resource.Quantity) bool {
	for i, input := range device.Spec.Inputs {
		if input.Name == inputName {
			device.Spec.Inputs[i].Value = value
			return true
		}
	}
	return false
}