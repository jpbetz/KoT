package main

import (
	"context"
	"fmt"
	"math"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	deepseav1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"
	v1 "github.com/jpbetz/KoT/apis/things/v1"
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
)

// PressureController reconciles a Device object
type PressureController struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=things.kubecon.io,resources=devices,verbs=get;list;watch;create;update;patch;delete

func (r *PressureController) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("device", req.NamespacedName)

	var device v1.Device
	if err := r.Get(ctx, req.NamespacedName, &device); err != nil {
		log.Error(err, "Failed to fetch Device")
		return ctrl.Result{}, ignoreNotFound(err)
	}

	if _, ok := getOutputV1(&device, "pressure"); ok {
		// reconcile pressure by activating pumps as needed
		err := r.ReconcilePressure(device)
		if err != nil {
			log.Error(err, "Failed to reconcile pressure")
		}
	}
	return ctrl.Result{}, nil
}

func (r *PressureController) ReconcilePressure(pressureDevice v1.Device) error {
	ctx := context.Background()

	// find the module this device belongs to, so we can find the pumps device
	if moduleName, ok := pressureDevice.Labels["module"]; ok {
		var m deepseav1alpha1.Module
		err := r.Get(ctx, types.NamespacedName{Namespace: pressureDevice.Namespace, Name: moduleName}, &m)
		if err != nil {
			return err
		}

		// find the pump device
		var pumpDevice v1.Device
		err = r.Get(ctx, types.NamespacedName{Namespace: m.Namespace, Name: m.Spec.Devices.Pump}, &pumpDevice)
		if err != nil {
			return err
		}

		// find the inputs and outputs we need to reconcile pressure by activating pumps
		pump, ok := getInputV1(&pumpDevice, "activeCount")
		if !ok {
			return fmt.Errorf("unable to find pump input: %s", pumpDevice.Name)
		}
		pressure, ok := getOutputV1(&pressureDevice, "pressure")
		if !ok {
			return fmt.Errorf("unable to find pressure output")
		}

		// calculate how many pumps to activate
		currentPressure := float64(pressure.Float.MilliValue()) / 1000
		activePumps := calculateActivePumps(currentPressure)
		if activePumps != nil {
			pump.Integer = activePumps
			err = r.Update(ctx, &pumpDevice)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func calculateActivePumps(pressure float64) *int32 {
	// We have 5 pumps, and don't want the pressure to exceed 0.5 in either direction
	count := 3 + (10.0-pressure)*(3/0.5)
	count = math.Max(count, 0)
	count = math.Min(count, 6)
	count = math.Round(count)
	result := int32(count)
	return &result
}

func (r *PressureController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Device{}).
		Complete(r)
}
