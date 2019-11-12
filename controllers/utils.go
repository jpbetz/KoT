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
	apierrs "k8s.io/apimachinery/pkg/api/errors"

	v1 "github.com/jpbetz/KoT/apis/things/v1"
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
)

func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

func getOutput(device v1alpha1.Device, outputName string) (*v1alpha1.Value, bool) {
	for i, output := range device.Status.Outputs {
		if output.Name == outputName {
			return &device.Status.Outputs[i], true
		}
	}
	return nil, false
}

func getInput(device v1alpha1.Device, inputName string) (*v1alpha1.Value, bool) {
	for i, input := range device.Spec.Inputs {
		if input.Name == inputName {
			return &device.Spec.Inputs[i], true
		}
	}
	return nil, false
}

func getOutputV1(device v1.Device, outputName string) (*v1.Value, bool) {
	for i, output := range device.Status.Outputs {
		if output.Name == outputName {
			return &device.Status.Outputs[i], true
		}
	}
	return nil, false
}

func getInputV1(device v1.Device, inputName string) (*v1.Value, bool) {
	for i, input := range device.Spec.Inputs {
		if input.Name == inputName {
			return &device.Spec.Inputs[i], true
		}
	}
	return nil, false
}
