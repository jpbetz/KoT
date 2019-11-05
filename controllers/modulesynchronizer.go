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
	deepseav1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	simulatorclient "github.com/jpbetz/KoT/simulator/service/client"
)

type moduleSynchronizer struct {
	client.Client
	SimulatorClient *simulatorclient.Client
	Log logr.Logger
}

// Start checks the state of the simulated modules against the module resources.
// - Registers simulated modules for any module resources that do not already have a corresponding
//   simulated module.
func (s *moduleSynchronizer) Start(stopCh <-chan struct{}) error {
	fn := func() {
		ctx := context.Background()
		var list deepseav1alpha1.ModuleList
		err := s.List(ctx, &list)
		if err != nil {
			s.Log.Error(err, "Failed list devices")
			return
		}
		for _, moduleObj := range list.Items {
			select {
			case <-stopCh:
				return
			default:
			}
			log := s.Log.WithValues("module", moduleObj.Name)
			module, err := s.SimulatorClient.GetModule(moduleObj.Name)
			if err != nil {
				log.Error(err, "Failed get corresponding simulator module")
				continue
			}
			if module == nil {
				log.Info("Module not found in simulator, registering it")
				err = s.SimulatorClient.PutModule(&moduleObj)
				if err != nil {
					log.Error(err, "Failed update simulator module")
				}
				continue
			}
		}
	}
	wait.Until(fn, time.Second * 5, stopCh)
	return nil
}
