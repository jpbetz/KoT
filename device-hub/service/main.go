package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"jpbetz/KoT/device-hub/service/types"
)

func main() {
	s := &server{}
	s.deviceHub = &types.DeviceHub{
		Devices: []*types.Device{
			{
				ID: "device1",
				Outputs: []*types.Output{
					{ID: "slider", Value: 85.0},
					{ID: "switch", Value: 1.0},
				},
				Inputs: []*types.Input{
					{ID: "value", Value: 85.0},
					{ID: "light", Value: 1.0},
				},
			},
			{
				ID: "device2",
				Outputs: []*types.Output{
					{ID: "slider", Value: 85.0},
					{ID: "switch", Value: 1.0},
				},
				Inputs: []*types.Input{
					{ID: "value", Value: 85.0},
					{ID: "light", Value: 1.0},
				},
			},
		},
	}
	s.websockets = newWebsocketManager()
	go s.websockets.run()
	router := mux.NewRouter()
	router.HandleFunc("/api/", s.deviceHubHandler)
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
	deviceHub  *types.DeviceHub
	websockets *WebsocketManager
}

func (s *server) deviceHubHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	switch r.Method {
	case "GET":
		s.mu.Lock()
		defer s.mu.Unlock()

		js, err := json.Marshal(s.deviceHub)
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
	
func (s *server) inputHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	inputID := vars["inputID"]

	s.mu.Lock()
	defer s.mu.Unlock()
	device := s.deviceHub.GetDevice(deviceID)
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
			Path: deviceID + "." + inputID,
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
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	outputID := vars["outputID"]

	s.mu.Lock()
	defer s.mu.Unlock()
	device := s.deviceHub.GetDevice(deviceID)
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
			Path: deviceID + "." + outputID,
			Value: output.Value,
		}
		s.websockets.broadcast <-msg
		return
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}
