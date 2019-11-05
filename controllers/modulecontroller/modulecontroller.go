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

package modulecontroller

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-logr/logr"
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	deepseav1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"

	simulatorclient "github.com/jpbetz/KoT/simulator/service/client"
)

// ModuleReconciler reconciles a Module object
type ModuleReconciler struct {
	client.Client
	SimulatorClient *simulatorclient.Client
	Log logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=things.kubecon.io,resources=modules,verbs=get;list;watch;create;update;patch;delete

func (r *ModuleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("module", req.NamespacedName)

	var m deepseav1alpha1.Module
	if err := r.Get(ctx, req.NamespacedName, &m); err != nil {
		log.Error(err, "unable to fetch Module")
		return ctrl.Result{}, ignoreNotFound(err)
	}

	// Reconcile device owner references with this modules devices
	gvk := m.GroupVersionKind()
	ownerRef := v1.OwnerReference{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Name:       m.GetName(),
		UID:        m.GetUID(),
	}
	for _, deviceName := range []string{m.Spec.Devices.Pump, m.Spec.Devices.PressureSensor, m.Spec.Devices.WaterAlarm} {
		var device v1alpha1.Device
		err := r.Get(ctx, types.NamespacedName{Namespace: m.Namespace, Name: deviceName}, &device)
		if err != nil {
			return ctrl.Result{}, err
		}

		found := false
		for _, ref := range device.OwnerReferences {
			if equality.Semantic.DeepEqual(ownerRef, ref) {
				found = true
				break
			}
		}
		if !found {
			device.OwnerReferences = append(device.OwnerReferences, ownerRef)
			err = r.Update(ctx, &device)
		}
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

type syncRunnable struct {
	r *ModuleReconciler
}

func (s *syncRunnable) Start(stopCh <-chan struct{}) error {
	fn := func() {
		ctx := context.Background()
		var list deepseav1alpha1.ModuleList
		err := s.r.List(ctx, &list)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("failed list devices: %v", err))
			return
		}
		for _, m := range list.Items {
			select {
			case <-stopCh:
				return
			default:
			}
			module, err := s.r.SimulatorClient.GetModule(m.Name)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("failed get module while reconciling for %s: %v", m.Name, err))
				continue
			}
			if module == nil {
				s.r.Log.Info("Module not found in simulator, registering it", "module", m.Name)
				err = s.r.SimulatorClient.PutModule(&m)
				if err != nil {
					utilruntime.HandleError(fmt.Errorf("failed reconcile module for %s: %v", m.Name, err))
				}
				continue
			}
		}
	}
	wait.Until(fn, time.Second * 5, stopCh)
	return nil
}


func (r *ModuleReconciler) SyncModules() manager.Runnable {
	return &syncRunnable{r}
}


type simulationRunnable struct {
	r *ModuleReconciler
}

func (s *simulationRunnable) Start(stopCh <-chan struct{}) error {
	fn := func() {
		ctx := context.Background()
		var list deepseav1alpha1.ModuleList
		err := s.r.List(ctx, &list)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("failed to list devices: %v", err))
			return
		}
		for _, m := range list.Items {
			select {
			case <-stopCh:
				return
			default:
			}

			var pumpDevice v1alpha1.Device
			err := s.r.Get(ctx, types.NamespacedName{Namespace: m.Namespace, Name: m.Spec.Devices.Pump}, &pumpDevice)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("failed to get pump devices: %v", err))
				return
			}
			var pump v1alpha1.Value
			for _, i := range pumpDevice.Status.ObservedInputs {
				if i.Name == "activeCount" {
					pump = i
				}
			}

			var pressureDevice v1alpha1.Device
			err = s.r.Get(ctx, types.NamespacedName{Namespace: m.Namespace, Name: m.Spec.Devices.PressureSensor}, &pressureDevice)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("failed to get pressure devices: %v", err))
				return
			}
			var pressure v1alpha1.Value
			for _, i := range pressureDevice.Status.Outputs {
				if i.Name == "pressure" {
					pressure = i
				}
			}

			freq := 1e-10 // ~1 minute period
			amp := 200.0  // ~4 bars amplitude
			n := time.Now().UnixNano()
			simPressureChange := math.Sin(freq*float64(n)) * amp

			// TODO: if pumps are active account for their effect

			// problem is the pump is on a different device than the current one
			// we should move this entire loop over to module, and lookup the devices for each module

			pumpVal := float64(pump.Value.Value())
			pumpPressureChange := (pumpVal - 2.5) * 100

			change := simPressureChange + pumpPressureChange
			deltaQuantity := resource.NewMilliQuantity(int64(change), resource.DecimalExponent)
			pressure.Value.Add(*deltaQuantity)

			err = s.r.SimulatorClient.PutOutput(pressureDevice.Name, pressure.Name, &pressure)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("failed to update device input for simulation: %v", err))
				continue
			}
		}
	}
	wait.Until(fn, time.Millisecond * 500, stopCh)
	return nil
}

func (r *ModuleReconciler) SimulatePressureChanges() manager.Runnable {
	return &simulationRunnable{r}
}

func (r *ModuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&deepseav1alpha1.Module{}).
		Complete(r)
}

func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}