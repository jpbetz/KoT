package main

import (
	"flag"
	"net/http"
	"time"

	//kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/jpbetz/KoT/controller/pkg/devicehub"
	clientset "github.com/jpbetz/KoT/controller/pkg/generated/clientset/versioned"
	informers "github.com/jpbetz/KoT/controller/pkg/generated/informers/externalversions"

	"k8s.io/klog"
	"k8s.io/sample-controller/pkg/signals"
)

var (
	masterURL  string
	kubeconfig string
	deviceHubURL string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	deviceClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}

	//kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	deviceInformerFactory := informers.NewSharedInformerFactory(deviceClient, time.Second*30)

	dhClient := &devicehub.Client{Url: deviceHubURL, Client: &http.Client{}}
	controller := NewController(dhClient, kubeClient, deviceClient, deviceInformerFactory.K8softhings().V1alpha1().Devices())

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	//kubeInformerFactory.Start(stopCh)
	deviceInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&deviceHubURL, "devicehub", "", "The address of the device hub to connect to.")
}