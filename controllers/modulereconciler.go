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

	"github.com/go-logr/logr"
	deepseav1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

// Reconcile ensures that all devices referenced by the module have owner references back to the module.
func (r *ModuleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("module", req.NamespacedName)

	var m deepseav1alpha1.Module
	if err := r.Get(ctx, req.NamespacedName, &m); err != nil {
		log.Error(err, "unable to fetch Module")
		return ctrl.Result{}, ignoreNotFound(err)
	}

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
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *ModuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&deepseav1alpha1.Module{}).
		Complete(r)
}
