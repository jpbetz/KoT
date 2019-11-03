package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/jpbetz/KoT/device-hub/service/types"
)

func main() {
	s := &server{}
	s.modules = &types.Modules{
		Modules: []*types.Module{
			{
				ID: "command",
				PressureSensor: &types.Device{
					ID: "pressureSensor1",
					Outputs: []*types.Output{
						{ID: "pressure", Value: 10.0},
					},
				},
				WaterAlarm: &types.Device{
					ID: "alarm1",
					Outputs: []*types.Output{
						{ID: "alarm", Value: 0.0},
					},
				},
				Pump: &types.Device{
					ID: "pumps1",
					Inputs: []*types.Input{
						{ID: "activeCount", Value: 1.0},
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
		Addr: ":8080",
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
	input := device.GetInput(inputID)
	if input == nil {
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
		var updatedInput types.Input
		err := decoder.Decode(&updatedInput)
		if err != nil {
			log.Printf("error decoding request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		input.Value = updatedInput.Value
		msg := &types.ValueChangedMessage{
			Path: module.ID + "." + deviceID + "." + inputID,
			Value: input.Value,
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
	output := device.GetOutput(outputID)
	if output == nil {
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
		var updatedOutput types.Output
		err := decoder.Decode(&updatedOutput)
		if err != nil {
			log.Printf("error decoding request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		output.Value = updatedOutput.Value
		msg := &types.ValueChangedMessage{
			Path: module.ID + "." + deviceID + "." + outputID,
			Value: output.Value,
		}
		s.websockets.broadcast <-msg
		return
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}
