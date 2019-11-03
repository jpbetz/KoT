package devicecontroller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/jpbetz/KoT/apis/things/v1alpha1"
	simulatorclient "github.com/jpbetz/KoT/device-hub/service/client"
)

// DeviceReconciler reconciles a Device object
type DeviceReconciler struct {
	client.Client
	SimulatorClient *simulatorclient.Client
	Log logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=things.kubecon.com,resources=devices,verbs=get;list;watch;create;update;patch;delete

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

	// TODO: update the device if the spec has changed


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
			status, err := s.r.SimulatorClient.CheckDeviceStatus(d.Name)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("failed get device status for %s: %v", d.Name, err))
				continue
			}
			if !equality.Semantic.DeepEqual(d.Status, status) {
				d.Status = status
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


func (r *DeviceReconciler) SyncStatus() manager.Runnable {
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