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
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/klog"

	deepseav1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
)

type moduleHandler struct {
	s *server
}

func (m *moduleHandler) OnAdd(obj interface{}) {
	if module, ok := obj.(*deepseav1alpha1.Module); ok {
		m.s.websockets.SendModuleCreated(module.Name)
	}
}
func (m *moduleHandler) OnUpdate(oldObj, newObj interface{}) {
	if module, ok := newObj.(*deepseav1alpha1.Module); ok {
		m.s.websockets.SendModuleUpdated(module.Name)
	}
}
func (m *moduleHandler) OnDelete(obj interface{}) {
	if module, ok := obj.(*deepseav1alpha1.Module); ok {
		m.s.websockets.SendModuleDeleted(module.Name)
	}
}

type deviceHandler struct {
	s *server
}

func (d *deviceHandler) OnAdd(obj interface{}) {
	if device, ok := obj.(*v1alpha1.Device); ok {
		if module, ok := findModule(*device); ok {
			d.s.websockets.SendModuleUpdated(module)
		}
	}
}
func (d *deviceHandler) OnUpdate(oldObj, newObj interface{}) {
	oldDevice, oldOk := oldObj.(*v1alpha1.Device)
	newDevice, newOk := newObj.(*v1alpha1.Device)
	if newOk && oldOk {
		if module, ok := findModule(*newDevice); ok {
			d.onChangedValues(oldDevice.Spec.Inputs, newDevice.Spec.Inputs, func(input v1alpha1.Value) {
				err := d.s.patchObservedInput(&input, newDevice.Name)
				if err != nil {
					klog.Errorf("error patching observed input: %v", err)
				}
				d.s.websockets.SendValueChanged(module+"."+newDevice.Name+"."+input.Name, input.Value)
			})
			d.onChangedValues(oldDevice.Status.ObservedInputs, newDevice.Status.ObservedInputs, func(input v1alpha1.Value) {
				d.s.websockets.SendValueChanged(module+"."+newDevice.Name+"."+input.Name, input.Value)
			})
			d.onChangedValues(oldDevice.Status.Outputs, newDevice.Status.Outputs, func(output v1alpha1.Value) {
				d.s.websockets.SendValueChanged(module+"."+newDevice.Name+"."+output.Name, output.Value)
			})
		}
	}
}

func (d *deviceHandler) onChangedValues(oldValues, newValues []v1alpha1.Value, handler func(v1alpha1.Value)) {
	m := map[string]v1alpha1.Value{}
	for _, v := range oldValues {
		m[v.Name] = v
	}

	for _, newValue := range newValues {
		if oldValue, ok := m[newValue.Name]; ok {
			if !equality.Semantic.DeepEqual(oldValue, newValue) {
				handler(newValue)
			}
		}
	}
}

func (d *deviceHandler) OnDelete(obj interface{}) {
	if device, ok := obj.(*v1alpha1.Device); ok {
		if module, ok := findModule(*device); ok {
			d.s.websockets.SendModuleUpdated(module)
		}
	}
}

func findModule(d v1alpha1.Device) (string, bool) {
	moduleName, ok := d.Labels["module"]
	return moduleName, ok
}