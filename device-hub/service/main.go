package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/jpbetz/KoT/apis/things/v1alpha1"
	"github.com/jpbetz/KoT/device-hub/service/types"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8085", "The address to bind to.")
	flag.Parse()

	quantity := func(s string) resource.Quantity {
		q, err := resource.ParseQuantity(s)
		if err != nil {
			log.Fatalf("failed to parse quantity: %s", s)
		}
		return q
	}
	s := &server{}
	s.modules = &types.Modules{
		Modules: []*types.Module{
			{
				ID: "command",
				PressureSensor: &types.Device{
					ID: "pressureSensor1",
					Outputs: []v1alpha1.Value{
						{Name: "pressure", Type: "float", Value: quantity("10.0")},
					},
				},
				WaterAlarm: &types.Device{
					ID: "alarm1",
					Outputs: []v1alpha1.Value{
						{Name: "alarm", Type: "boolean", Value: quantity("0.0")},
					},
				},
				Pump: &types.Device{
					ID: "pumps1",
					Inputs: []v1alpha1.Value{
						{Name: "activeCount", Type: "integer", Value: quantity("1.0")},
					},
				},
			},
		},
	}
	s.websockets = newWebsocketManager()
	go s.websockets.run()
	router := mux.NewRouter()
	router.HandleFunc("/api/modules", s.modulesHandler)
	router.HandleFunc("/api/modules/{moduleID}", s.moduleHandler)
	router.HandleFunc("/api/devices/{deviceID}/inputs/{inputID}", s.inputHandler)
	router.HandleFunc("/api/devices/{deviceID}/outputs/{outputID}", s.outputHandler)
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocketHandler(s, w, r)
	})
	router.PathPrefix("/").HandlerFunc(staticContentHandler)
	server := &http.Server{
		Handler: router,
		Addr: addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout: 15 * time.Second,
	}
	log.Println("Server is starting")
	log.Fatal(server.ListenAndServe())
}

type server struct {
	mu         sync.Mutex
	modules    *types.Modules
	websockets *WebsocketManager
}

const (
	targetPressure = 10.0 // unit: bars, for depth of 90 meters
	lowPressure = 8.0
	emergencyPressure = 6.0
	highPressure = 12.0
	pressureDropPerSec = 0.05
)

func (s *server) modulesHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	switch r.Method {
	case "GET":
		s.mu.Lock()
		defer s.mu.Unlock()

		js, err := json.Marshal(s.modules)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			log.Printf("error: %v", err)
		}
		return
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}

func (s *server) moduleHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	vars := mux.Vars(r)
	moduleID := vars["moduleID"]

	s.mu.Lock()
	defer s.mu.Unlock()
	module := s.modules.GetModule(moduleID)
	switch r.Method {
	case "GET":
		if module == nil {
			http.NotFound(w, r)
			return
		}
		js, err := json.Marshal(module)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			log.Printf("error: %v", err)
		}
		return
	case "PUT":
		decoder := json.NewDecoder(r.Body)
		updatedModule := &types.Module{}
		err := decoder.Decode(updatedModule)
		if err != nil {
			log.Printf("error decoding request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if module != nil && module.ID != updatedModule.ID {
			http.Error(w, "ID in request does not match ID in path", http.StatusBadRequest)
		}
		s.modules.PutModule(updatedModule)
		return
	case "DELETE":
		if module == nil {
			http.NotFound(w, r)
			return
		}
		s.modules.DeleteModule(moduleID)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}
	
func (s *server) inputHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	inputID := vars["inputID"]

	s.mu.Lock()
	defer s.mu.Unlock()
	module, device := s.modules.GetDevice(deviceID)
	if device == nil {
		http.NotFound(w, r)
		return
	}
	input, ok := device.GetInput(inputID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	
	switch r.Method {
	case "GET":
		js, err := json.Marshal(input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(js)
		return
	case "PUT":
		decoder := json.NewDecoder(r.Body)
		var updatedInput v1alpha1.Value
		err := decoder.Decode(&updatedInput)
		if err != nil {
			log.Printf("error decoding request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		device.SetInput(input.Name, updatedInput.Value)
		msg := &types.ValueChangedMessage{
			Path: module.ID + "." + deviceID + "." + inputID,
			Value: updatedInput.Value,
		}
		s.websockets.broadcast <-msg
		return
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}

func (s *server) outputHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	outputID := vars["outputID"]

	s.mu.Lock()
	defer s.mu.Unlock()
	module, device := s.modules.GetDevice(deviceID)
	if device == nil {
		http.NotFound(w, r)
		return
	}
	output, ok := device.GetOutput(outputID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case "GET":
		js, err := json.Marshal(output)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(js)
		return
	case "PUT":
		decoder := json.NewDecoder(r.Body)
		var updatedOutput v1alpha1.Value
		err := decoder.Decode(&updatedOutput)
		if err != nil {
			log.Printf("error decoding request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		device.SetOutput(output.Name, updatedOutput.Value)
		msg := &types.ValueChangedMessage{
			Path: module.ID + "." + deviceID + "." + outputID,
			Value: updatedOutput.Value,
		}
		s.websockets.broadcast <-msg
		return
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}
