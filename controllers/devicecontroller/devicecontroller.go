package devicecontroller

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
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
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, ignoreNotFound(err)
	}

	// TODO: Add reconciliation logic here
	// err := r.SimulatorClient.PutDevice(&device)
	// if err != nil {
	// 	utilruntime.HandleError(fmt.Errorf("failed to update simulator device for %s: %v", device.Name, err))
	// }

	return ctrl.Result{}, nil
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

type simulationRunnable struct {
	r *DeviceReconciler
}

func (s *simulationRunnable) Start(stopCh <-chan struct{}) error {
	fn := func() {
		ctx := context.Background()
		var list v1alpha1.DeviceList
		err := s.r.List(ctx, &list)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("failed to list devices: %v", err))
			return
		}
		for _, d := range list.Items {
			select {
			case <-stopCh:
				return
			default:
			}
			for _, i := range d.Status.Outputs {
				if i.Name == "pressure" {
					freq := 1e-10 // ~1 minute period
					amp := 200.0 // ~4 bars amplitude
					n := time.Now().UnixNano()
					delta := math.Sin(freq*float64(n))*amp
					deltaQuantity := resource.NewMilliQuantity(int64(delta), resource.DecimalExponent)
					i.Value.Add(*deltaQuantity)
					err = s.r.SimulatorClient.PutOutput(d.Name, i.Name, &i)
					if err != nil {
						utilruntime.HandleError(fmt.Errorf("failed to update device input for simulation: %v", err))
						continue
					}
				}
			}
		}
	}
	wait.Until(fn, time.Millisecond * 500, stopCh)
	return nil
}

func (r *DeviceReconciler) SimulatePressureChanges() manager.Runnable {
	return &simulationRunnable{r}
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