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
	"math"
	"math/rand"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/jpbetz/KoT/apis/things/v1alpha1"

	deepseav1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"
)

type SimulationRunner struct {
	client.Client
	Log logr.Logger
}

// Start simulates pressure changes are a result of the number of pumps active and environmental effects
func (s *SimulationRunner) Start(stopCh <-chan struct{}) error {
	fn := func() {
		select {
		case <-stopCh:
			return
		default:
		}
		ctx := context.Background()
		var list deepseav1alpha1.ModuleList
		err := s.List(ctx, &list)
		if err != nil {
			s.Log.Error(err, "failed to list devices")
			return
		}
		for _, m := range list.Items {
			log := s.Log.WithValues("module", m.Name)

			// Find both the pump and pressure device, and their inputs and outputs
			var pumpDevice v1alpha1.Device
			err := s.Get(ctx, types.NamespacedName{Namespace: m.Namespace, Name: m.Spec.Devices.Pump}, &pumpDevice)
			if err != nil {
				log.Error(err, "Failed to get pump device for module")
				return
			}
			pump, ok := getInput(&pumpDevice, "activeCount")
			if !ok {
				log.Error(err, "Failed to find pump input")
				return
			}
			var pressureDevice v1alpha1.Device
			err = s.Get(ctx, types.NamespacedName{Namespace: m.Namespace, Name: m.Spec.Devices.PressureSensor}, &pressureDevice)
			if err != nil {
				log.Error(err, "Failed to get pressure device")
				return
			}

			change := calculatePressureChange(pump.Value.Value())

			var pressure *v1alpha1.Value
			pressure, ok = getOutput(&pressureDevice, "pressure")
			if !ok {
				value := resource.NewQuantity(10, resource.DecimalSI)
				pressure = &v1alpha1.Value{Name: "pressure", Value: *value, Type: v1alpha1.FloatType}
				pressureDevice.Status.Outputs = append(pressureDevice.Status.Outputs, *pressure)
			}
			quantity := &pressure.Value
			changeQuantity := resource.NewMilliQuantity(int64(change), resource.DecimalExponent)
			quantity.Add(*changeQuantity)

			// Write simulated pressure to device
			err = s.Client.Update(ctx, &pressureDevice)
			if err != nil {
				log.Error(err, "Failed to update pressure device")
				return
			}
		}
	}
	wait.Until(fn, time.Millisecond*500, stopCh)
	return nil
}

func calculatePressureChange(pumpsActive int64) float64 {
	// Calculate the simulated net change in pressure for the current iteration

	// Use a wave to simulate environmental pressure changes.
	freq := 1e-10 // ~1 minute period
	amp := 200.0  // ~4 bars amplitude
	n := time.Now().UnixNano() + int64(rand.Intn(1000)-500)
	simPressureChange := math.Sin(freq*float64(n)) * amp

	// 2.5 pumps are required for equilibrium
	// calculate pressure change if pumps are not at equilibrium
	pumpVal := float64(pumpsActive)
	pumpPressureChange := (pumpVal - 3.0) * 100

	return simPressureChange + pumpPressureChange
}
