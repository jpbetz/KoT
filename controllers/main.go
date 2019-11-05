package main

import (
	"flag"
	"net/http"
	"os"

	deepseav1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	// +kubebuilder:scaffold:imports

	"github.com/jpbetz/KoT/apis/things/v1alpha1"

	"github.com/jpbetz/KoT/controllers/devicecontroller"
	"github.com/jpbetz/KoT/controllers/modulecontroller"
	"github.com/jpbetz/KoT/simulator/service/client"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	_ = deepseav1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var simulatorAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&simulatorAddr, "simulator-addr", ":8085", "The address of the device simulator service.")
	flag.Parse()

	ctrl.SetLogger(klogr.New())

	simulatorClient := &client.Client{Url: simulatorAddr, Client: &http.Client{}}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{Scheme: scheme, MetricsBindAddress: metricsAddr})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	deviceReconciler := &devicecontroller.DeviceReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Captain"),
		SimulatorClient: simulatorClient,
		Scheme: mgr.GetScheme(),
	}
	if err = deviceReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Captain")
		os.Exit(1)
	}
	if err = mgr.Add(deviceReconciler.SyncDevices()); err != nil {
		setupLog.Error(err, "unable to add sync status runnable", "controller", "Captain")
		os.Exit(1)
	}

	moduleReconciler := &modulecontroller.ModuleReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Captain"),
		SimulatorClient: simulatorClient,
		Scheme: mgr.GetScheme(),
	}
	if err = moduleReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Captain")
		os.Exit(1)
	}
	if err = mgr.Add(moduleReconciler.SyncModules()); err != nil {
		setupLog.Error(err, "unable to add sync status runnable", "controller", "Captain")
		os.Exit(1)
	}
	if err = mgr.Add(moduleReconciler.SimulatePressureChanges()); err != nil {
		setupLog.Error(err, "unable to add simulation runnable", "controller", "Captain")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
