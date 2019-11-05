package main

import (
	"context"
	"fmt"
	"math"

	"github.com/go-logr/logr"
	deepseav1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	simulatorclient "github.com/jpbetz/KoT/simulator/service/client"
)

// DeviceReconciler reconciles a Device object
type DeviceReconciler struct {
	client.Client
	SimulatorClient *simulatorclient.Client
	Log logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=things.kubecon.io,resources=devices,verbs=get;list;watch;create;update;patch;delete

func (r *DeviceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("device", req.NamespacedName)

	var device v1alpha1.Device
	if err := r.Get(ctx, req.NamespacedName, &device); err != nil {
		log.Error(err, "Failed to fetch Device")
		return ctrl.Result{}, ignoreNotFound(err)
	}

	// reconcile pressure by activating pumps as needed
	if _, ok := getOutput(device, "pressure"); ok {
		err := r.ReconcilePressure(device)
		if err != nil {
			log.Error(err, "Failed to reconcile pressure")
		}
	}

	// reconcile device inputs
	// if an observed input in status differs than the desired input in spec, update the device with
	// the desired inputs
	if !equality.Semantic.DeepEqual(device.Spec.Inputs, device.Status.ObservedInputs) {
		// only sets spec field, which in this case, is the inputs
		err := r.SimulatorClient.PutDevice(&device)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

const desiredPressure = 10.0

func (r *DeviceReconciler) ReconcilePressure(pressureDevice v1alpha1.Device) error {
	ctx := context.Background()

	// find the module this device belongs to, so we can find the pumps device
	var moduleName string
	for _, ref := range pressureDevice.OwnerReferences {
		if ref.Kind == "Module" && ref.APIVersion == deepseav1alpha1.SchemeGroupVersion.String() {
			moduleName = ref.Name

			var m deepseav1alpha1.Module
			err := r.Get(ctx, types.NamespacedName{Namespace: pressureDevice.Namespace, Name: moduleName}, &m)
			if err != nil {
				return err
			}

			// find the pump device
			var pumpDevice v1alpha1.Device
			err = r.Get(ctx, types.NamespacedName{Namespace: m.Namespace, Name: m.Spec.Devices.Pump}, &pumpDevice)
			if err != nil {
				return err
			}

			// find the inputs and outputs we need to reconcile pressure by activating pumps
			pump, ok := getInput(pumpDevice, "activeCount")
			if !ok {
				return fmt.Errorf("unable to find pump input: %s", pumpDevice.Name)
			}
			pressure, ok := getOutput(pressureDevice, "pressure")
			if !ok {
				return fmt.Errorf("unable to find pressure output")
			}

			// calculate how many pumps to activate
			currentPressure := float64(pressure.Value.MilliValue()) / 1000
			activePumps := calculateActivePumps(currentPressure)
			pump.Value.Set(activePumps)

			// activate the pumps
			if !setInputValue(pumpDevice, pump.Name, pump.Value) {
				return fmt.Errorf("unable to find pump input: %s", pumpDevice.Name)
			}
			err = r.Update(ctx, &pumpDevice)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func calculateActivePumps(pressure float64) int64 {
	// We have 5 pumps, and don't want the pressure to exceed 0.5 in either direction
	count := 2.5 + (desiredPressure - pressure)*(2.5/0.5)
	count = math.Max(count, 0)
	count = math.Min(count, 5)
	count = math.Round(count)
	return int64(count)
}

func (r *DeviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Device{}).
		Complete(r)
}