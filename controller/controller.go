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
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	"github.com/jpbetz/KoT/controller/pkg/devicehub"
	clientset "github.com/jpbetz/KoT/controller/pkg/generated/clientset/versioned"
	devicescheme "github.com/jpbetz/KoT/controller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/jpbetz/KoT/controller/pkg/generated/informers/externalversions/k8softhings/v1alpha1"
	listers "github.com/jpbetz/KoT/controller/pkg/generated/listers/k8softhings/v1alpha1"
)

const controllerAgentName = "device-controller"

type Controller struct {
	devicehubclient *devicehub.Client
	kubeclientset kubernetes.Interface
	deviceclientset clientset.Interface

	deviceLister        listers.DeviceLister
	deviceSynced        cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func NewController(
	devicehubclient *devicehub.Client,
	kubeclientset kubernetes.Interface,
	deviceclientset clientset.Interface,
	deviceInformer informers.DeviceInformer) *Controller {

	utilruntime.Must(devicescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		devicehubclient: devicehubclient,
		kubeclientset: kubeclientset,
		deviceclientset: deviceclientset,
		deviceLister:        deviceInformer.Lister(),
		deviceSynced:        deviceInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Devices"),
		recorder:          recorder,
	}

	klog.Info("Setting up event handlers")
	deviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueDevice,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueDevice(new)
		},
	})

	return controller
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Device resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Device resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Device resource with this namespace/name
	device, err := c.deviceLister.Devices(namespace).Get(name)
	if err != nil {
		// The Device resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("device '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	// Apply logic to the outputs based on the inputs
	for _, i := range device.Spec.Inputs {
		if i.Name == "value" {
			for _, o := range device.Status.Outputs {
				if o.Name == "switch" {
					if o.Value == 0 {
						i.Value = 0
						break
					}
				}
				if o.Name == "slider" {
					if i.Value < o.Value {
						i.Value = o.Value
					}
				}
			}
		}
	}

	// reconcile differences between spec and status
	for _, specInput := range device.Spec.Inputs {
		for _, statusInput := range device.Status.Inputs {
			if specInput.Name == statusInput.Name &&  specInput.Value != statusInput.Value {
				klog.Infof("reconciling %s.%s spec: %f, status: %f", device.Name, specInput.Name, specInput.Value, statusInput.Value)
				err = c.devicehubclient.SetDeviceInput(device.Name, specInput.Name, specInput.Value)
				if err != nil {
					utilruntime.HandleError(fmt.Errorf("failed to set device %s specInput %s: %v", device.Name, specInput.Name, err))
				}
			}
		}
	}
	return nil
}

func (c *Controller) syncDeviceHubData() {
	deviceHub, err := c.devicehubclient.CheckDeviceStatuses()
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to fetch device updates: %v", err))
		return
	}
	for _, deviceInfo := range deviceHub.Devices {
		device, err := c.deviceLister.Devices("default").Get(deviceInfo.ID)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("failed to find CR for device %s: %v", deviceInfo.ID, err))
			return
		}
		actualStatus := devicehub.ToDeviceStatus(deviceInfo)
		if equality.Semantic.DeepEqual(actualStatus, device.Status) == false {
			klog.Infof("Status of device changed, updating CR spec: %s\n%s", deviceInfo.ID, cmp.Diff(device.Status, actualStatus))

			device.Status = actualStatus
			device, err = c.deviceclientset.K8softhingsV1alpha1().Devices("default").Update(device)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("failed to update status of device %s/%s: %v", device.Namespace, device.Name, err))
				return
			}
		}
	}
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Device controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deviceSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Device resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	go wait.Until(c.syncDeviceHubData, time.Millisecond * 100, stopCh)

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// enqueueDevice takes a Device resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueDevice(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}