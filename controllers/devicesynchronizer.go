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
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	simulatorclient "github.com/jpbetz/KoT/simulator/service/client"
)

type deviceSynchronizer struct {
	client.Client
	SimulatorClient *simulatorclient.Client
	Log logr.Logger
}

// Start checks the state of the simulated devices against the device resources.
// - Registers simulated devices for any device resources that do not already have a corresponding
//   simulated device.
// - Writes the status of the simulated devices to the device resources they differ.
func (s *deviceSynchronizer) Start(stopCh <-chan struct{}) error {
	fn := func() {
		ctx := context.Background()
		var list v1alpha1.DeviceList
		err := s.List(ctx, &list)
		if err != nil {
			s.Log.Error(err, "Failed list devices")
			return
		}
		for _, deviceObj := range list.Items {
			select {
			case <-stopCh:
				return
			default:
			}
			log := s.Log.WithValues("device", deviceObj.Name)

			device, err := s.SimulatorClient.GetDevice(deviceObj.Name)
			if err != nil {
				log.Error(err, "failed to get corresponding simulator device")
				continue
			}
			if device == nil {
				s.Log.Info("Device not found in simulator, registering it", "device", deviceObj.Name)
				err = s.SimulatorClient.PutDevice(&deviceObj)
				if err != nil {
					log.Error(err,"failed to update simulator device")
				}
				continue
			}
			if !equality.Semantic.DeepEqual(deviceObj.Status, device.Status) {
				deviceObj.Status = device.Status
				err = s.Update(ctx, &deviceObj)
				if err != nil {
					log.Error(err,"failed update device status")
					continue
				}
			}
		}
	}
	wait.Until(fn, time.Millisecond * 100, stopCh)
	return nil
}
