package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	clientset "github.com/jpbetz/KoT/generated/clientset/versioned"
	informers "github.com/jpbetz/KoT/generated/informers/externalversions"
	deepsealisters "github.com/jpbetz/KoT/generated/listers/deepsea/v1alpha1"
	thingslisters "github.com/jpbetz/KoT/generated/listers/things/v1alpha1"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/runtime"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	deepseav1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
)


var (
	scheme   = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	_ = deepseav1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type server struct {
	mu         sync.Mutex
	websockets *WebsocketManager
	moduleLister deepsealisters.ModuleLister
	deviceLister thingslisters.DeviceLister
	client *clientset.Clientset

}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&addr, "addr", ":8085", "The address to bind to.")
}

var (
	kubeconfig string
	masterURL  string
	addr string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := make(chan struct{})//signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}

	informerFactory := informers.NewSharedInformerFactory(client, time.Second*30)
	modulesInformer := informerFactory.Deepsea().V1alpha1().Modules()
	devicesInformer := informerFactory.Things().V1alpha1().Devices()

	s := &server{
		websockets: newWebsocketManager(),
		moduleLister: modulesInformer.Lister(),
		deviceLister: devicesInformer.Lister(),
		client: client,
	}

	modulesInformer.Informer().AddEventHandler(&moduleHandler{s})
	devicesInformer.Informer().AddEventHandler(&deviceHandler{s})

	go s.websockets.run(stopCh)
	informerFactory.Start(stopCh)
	s.runServer(addr, stopCh)
}

func (s *server) runServer(addr string, stopCh <-chan struct{}) {
	router := mux.NewRouter()
	router.HandleFunc("/api/", s.everythingHandler)
	router.HandleFunc("/api/devices/{deviceName}/inputs/{inputName}", s.inputHandler)
	router.HandleFunc("/api/devices/{deviceName}/outputs/{outputName}", s.outputHandler)
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocketHandler(s, w, r)
	})
	router.PathPrefix("/").HandlerFunc(staticContentHandler)
	server := &http.Server{
		Handler:      router,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Server is starting")
	log.Fatal(server.ListenAndServe())
}

func (s *server) everythingHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	switch r.Method {
	case "GET":
		s.mu.Lock()
		defer s.mu.Unlock()

		modules, err := s.moduleLister.List(labels.Everything())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		devices, err := s.deviceLister.List(labels.Everything())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sendResponse(w, r, map[string]interface{}{"modules": modules, "devices": devices}, true)
		return
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}

func (s *server) inputHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	vars := mux.Vars(r)
	deviceName := vars["deviceName"]
	inputName := vars["inputName"]
	
	switch r.Method {
	case "PUT":
		valueIn := &v1alpha1.Value{}
		applyPut(w, r, valueIn, func() error {
			valueIn.Name = inputName
			err := s.patchInput(valueIn, deviceName)
			return err
		})
		return
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}

func (s *server) outputHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	vars := mux.Vars(r)
	deviceName := vars["deviceName"]
	outputName := vars["outputName"]

	switch r.Method {
	case "PUT":
		valueIn := &v1alpha1.Value{}
		applyPut(w, r, valueIn, func() error {
			valueIn.Name = outputName
			err := s.patchOutput(valueIn, deviceName)
			return err
		})
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}

func (s *server) patchInput(valueIn *v1alpha1.Value, deviceName string) error {
	input := `{"name": "` + valueIn.Name + `", "value": "` + valueIn.Value.String() + `"}`
	patch := `{"spec": {"inputs": [` + input + `]}, "status": {"observedInputs": [` + input + `]}}`
	_, err := s.client.ThingsV1alpha1().Devices("default").Patch(deviceName, types.MergePatchType, []byte(patch))
	return err
}

func (s *server) patchObservedInput(valueIn *v1alpha1.Value, deviceName string) error {
	input := `{"name": "` + valueIn.Name + `", "value": "` + valueIn.Value.String() + `"}`
	patch := `{"status": {"observedInputs": [` + input + `]}}`
	_, err := s.client.ThingsV1alpha1().Devices("default").Patch(deviceName, types.MergePatchType, []byte(patch))
	return err
}

func (s *server) patchOutput(valueIn *v1alpha1.Value, deviceName string) error {
	output := `{"name": "` + valueIn.Name + `", "value": "` + valueIn.Value.String() + `"}`
	patch := `{"status": {"outputs": [` + output + `]}}`
	_, err := s.client.ThingsV1alpha1().Devices("default").Patch(deviceName, types.MergePatchType, []byte(patch))
	return err
}

func applyPut(w http.ResponseWriter, r *http.Request, out interface{}, fn func() error) {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(out)
	if err != nil {
		log.Printf("error decoding request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = fn()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func sendResponse(w http.ResponseWriter, r *http.Request, obj interface{}, exists bool) {
	if !exists {
		http.NotFound(w, r)
		return
	}
	js, err := json.Marshal(obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		log.Printf("error: %v", err)
	}
}