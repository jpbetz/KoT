package devicecontroller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	deepseav1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/jpbetz/KoT/apis/things/v1alpha1"

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
		log.Error(err, "unable to fetch Device")
		return ctrl.Result{}, ignoreNotFound(err)
	}

	// reconcile which pumps are turned on to keep pressure at desired level
	isPressureDevice := false
	for _, i := range device.Status.Outputs {
		if i.Name == "pressure" {
			isPressureDevice = true
		}
	}
	if isPressureDevice {
		err := r.ReconcilePressure(device)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// reconcile status and spec
	if !equality.Semantic.DeepEqual(device.Spec.Inputs, device.Status.ObservedInputs) {
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

	// find the module this device belongs to
	var moduleName string
	for _, ref := range pressureDevice.OwnerReferences {
		if ref.Kind == "Module" && ref.APIVersion == deepseav1alpha1.SchemeGroupVersion.String() {
			moduleName = ref.Name
			break
		}
	}

	if moduleName == "" {
		// need the module to activate pumps to manager pressure, can't do anything without it
		return nil
	}
	var m deepseav1alpha1.Module
	err := r.Get(ctx, types.NamespacedName{Namespace: pressureDevice.Namespace, Name: moduleName}, &m)
	if err != nil {
		return err
	}

	var pumpDevice v1alpha1.Device
	err = r.Get(ctx, types.NamespacedName{Namespace: m.Namespace, Name: m.Spec.Devices.Pump}, &pumpDevice)
	if err != nil {
		return err
	}
	var pump v1alpha1.Value
	for _, i := range pumpDevice.Status.ObservedInputs {
		if i.Name == "activeCount" {
			pump = i
		}
	}

	var pressure v1alpha1.Value
	for _, i := range pressureDevice.Status.Outputs {
		if i.Name == "pressure" {
			pressure = i
		}
	}

	v := float64(pressure.Value.MilliValue()) / 1000
	var pumps int64
	switch {
	case v > desiredPressure+1:
		pumps = 0
	case v > desiredPressure + 0.5:
		pumps = 1
	case v > desiredPressure + 0.1:
		pumps = 2
	case v < desiredPressure - 1:
		pumps = 5
	case v < desiredPressure - 0.5:
		pumps = 4
	case v < desiredPressure - 0.1:
		pumps = 3
	}
	pump.Value = *resource.NewQuantity(pumps, resource.DecimalSI)
	err = r.SimulatorClient.PutInput(pumpDevice.Name, pump.Name, &pump)
	if err != nil {
		return err
	}
	return nil
}

type syncRunnable struct {
	r *DeviceReconciler
}

func (s *syncRunnable) Start(stopCh <-chan struct{}) error {
	fn := func() {
		ctx := context.Background()
		var list v1alpha1.DeviceList
		err := s.r.List(ctx, &list)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("failed list devices: %v", err))
			return
		}
		for _, d := range list.Items {
			select {
			case <-stopCh:
				return
			default:
			}
			device, err := s.r.SimulatorClient.GetDevice(d.Name)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("failed to get device for %s: %v", d.Name, err))
				continue
			}
			if device == nil {
				s.r.Log.Info("Device not found in simulator, registering it", "device", d.Name)
				err = s.r.SimulatorClient.PutDevice(&d)
				if err != nil {
					utilruntime.HandleError(fmt.Errorf("failed reconcile device for %s: %v", d.Name, err))
				}
				continue
			}
			if !equality.Semantic.DeepEqual(d.Status, device.Status) {
				d.Status = device.Status
				err = s.r.Update(ctx, &d)
				if err != nil {
					utilruntime.HandleError(fmt.Errorf("failed update device status for %s: %v", d.Name, err))
					continue
				}
			}
		}
	}
	wait.Until(fn, time.Millisecond * 100, stopCh)
	return nil
}


func (r *DeviceReconciler) SyncDevices() manager.Runnable {
	return &syncRunnable{r}
}

func (r *DeviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Device{}).
		Complete(r)
}

func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}