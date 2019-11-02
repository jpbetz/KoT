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
	s.simulatedDevices = &types.SimulatedDevices{
		Devices: []*types.Device{
			{
				ID: "command",
				Outputs: []*types.Output{
					{ID: "pressure", Value: 10.0},
					{ID: "waterSensor", Value: 0.0},
				},
				Inputs: []*types.Input{
					{ID: "pumpsActive", Value: 1.0},
					{ID: "alarm", Value: 0.0},
				},
			},
			{
				ID: "crew",
				Outputs: []*types.Output{
					{ID: "pressure", Value: 10.0},
					{ID: "waterSensor", Value: 0.0},
				},
				Inputs: []*types.Input{
					{ID: "pumpsActive", Value: 1.0},
					{ID: "alarm", Value: 0.0},
				},
			},
			{
				ID: "research",
				Outputs: []*types.Output{
					{ID: "pressure", Value: 10.0},
					{ID: "waterSensor", Value: 0.0},
				},
				Inputs: []*types.Input{
					{ID: "pumpsActive", Value: 1.0},
					{ID: "alarm", Value: 0.0},
				},
			},
		},
	}
	s.websockets = newWebsocketManager()
	go s.websockets.run()
	go s.simulate()
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
	mu               sync.Mutex
	simulatedDevices *types.SimulatedDevices
	websockets       *WebsocketManager
}

const (
	targetPressure = 10.0 // unit: bars, for depth of 90 meters
	lowPressure = 8.0
	emergencyPressure = 6.0
	highPressure = 12.0
	pressureDropPerSec = 0.05
)
func (s *server) simulate() {
	for range time.Tick(50 * time.Millisecond) {
		for _, d := range s.simulatedDevices.Devices {
			func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				pressure := d.GetOutput("pressure")
				active := d.GetInput("pumpsActive").Value
				if active > 3 {
					active = 3
				}
				if active < 0.9 || active > 1.9 {
					if active < 0.9 {
						pressure.Value -= 0.1 / 20.0
					}
					if active > 1.9 {
						pressure.Value += 0.05 * active / 20.0
					}

					msg := &types.ValueChangedMessage{
						Path: d.ID + "." + pressure.ID,
						Value: pressure.Value,
					}
					s.websockets.broadcast <-msg
				}
			}()
		}
	}
}

func (s *server) deviceHubHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	switch r.Method {
	case "GET":
		s.mu.Lock()
		defer s.mu.Unlock()

		js, err := json.Marshal(s.simulatedDevices)
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
	log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	inputID := vars["inputID"]

	s.mu.Lock()
	defer s.mu.Unlock()
	device := s.simulatedDevices.GetDevice(deviceID)
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
	log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	outputID := vars["outputID"]

	s.mu.Lock()
	defer s.mu.Unlock()
	device := s.simulatedDevices.GetDevice(deviceID)
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
