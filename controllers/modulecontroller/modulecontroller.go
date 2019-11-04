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
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
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

	var module deepseav1alpha1.Module
	if err := r.Get(ctx, req.NamespacedName, &module); err != nil {
		log.Error(err, "unable to fetch Module")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, ignoreNotFound(err)
	}

	// TODO: Add reconciliation logic here
	// err := r.SimulatorClient.PutModule(&module)
	// if err != nil {
	// 	utilruntime.HandleError(fmt.Errorf("failed to update simulator module for %s: %v", module.Name, err))
	// }

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