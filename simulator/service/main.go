package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/api/resource"

	deepseav1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
)

type server struct {
	mu         sync.Mutex
	modules    map[string]*deepseav1alpha1.Module
	devices    map[string]*v1alpha1.Device
	deviceModules map[string]string
	websockets *WebsocketManager
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8085", "The address to bind to.")
	flag.Parse()

	s := &server{}
	s.modules =  map[string]*deepseav1alpha1.Module{}
	s.devices = map[string]*v1alpha1.Device{}
	s.deviceModules = map[string]string{}

	s.websockets = newWebsocketManager()
	go s.websockets.run()
	router := mux.NewRouter()
	router.HandleFunc("/api/", s.everythingHandler)
	router.HandleFunc("/api/modules", s.modulesHandler)
	router.HandleFunc("/api/modules/{moduleID}", s.moduleHandler)
	router.HandleFunc("/api/devices", s.devicesHandler)
	router.HandleFunc("/api/devices/{deviceID}", s.deviceHandler)
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

func (s *server) everythingHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	switch r.Method {
	case "GET":
		s.mu.Lock()
		defer s.mu.Unlock()

		getResp(w, r, map[string]interface{}{"modules": s.modules, "devices": s.devices}, true)
		return
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}

func (s *server) modulesHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	switch r.Method {
	case "GET":
		s.mu.Lock()
		defer s.mu.Unlock()

		getResp(w, r, s.modules, true)
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
	module, exists := s.modules[moduleID]
	switch r.Method {
	case "GET":
		getResp(w, r, module, exists)
		return
	case "PUT":
		updatedModule := &deepseav1alpha1.Module{}
		applyPut(w, r, updatedModule, func() error {
			if exists && module.Name != updatedModule.Name {
				http.Error(w, "Name in request does not match Name in path", http.StatusBadRequest)
			}
			s.modules[moduleID] = updatedModule

			if exists {
				oldDevices := module.Spec.Devices
				for _, d := range []string{oldDevices.Pump, oldDevices.WaterAlarm, oldDevices.PressureSensor} {
					delete(s.deviceModules, d)
				}
			}
			newDevices := updatedModule.Spec.Devices
			for kind, deviceName := range map[string]string{"pump": newDevices.Pump, "alarm": newDevices.WaterAlarm, "pressure": newDevices.PressureSensor} {
				existingDevice, ok := s.devices[deviceName]
				if !ok {
					http.Error(w, fmt.Sprintf("Module %s references non-existent device %s", moduleID, deviceName), http.StatusBadRequest)
					return nil
				}

				if _, ok := s.deviceModules[deviceName]; !ok {
					// device is being registered
					switch kind {
					case "pump":
						existingDevice.Status = v1alpha1.DeviceStatus{
							ObservedInputs: []v1alpha1.Value{
								{Name: "activeCount", Type: v1alpha1.IntegerType, Value: quantity("1.0")},
							},
						}
					case "alarm":
						existingDevice.Status = v1alpha1.DeviceStatus{
							Outputs: []v1alpha1.Value{
								{Name: "alarm", Type: v1alpha1.BooleanType, Value: quantity("0.0")},
							},
						}
					case "pressure":
						existingDevice.Status = v1alpha1.DeviceStatus{
							Outputs: []v1alpha1.Value{
								{Name: "pressure", Type: v1alpha1.FloatType, Value: quantity("10.0")},
							},
						}
					default:
						http.Error(w, fmt.Sprintf("Unrecognized kind: %s", kind), http.StatusInternalServerError)
					}

				}
				s.deviceModules[deviceName] = updatedModule.Name
			}

			if !exists {
				s.websockets.SendModuleCreated(moduleID)
			}
			return nil
		})
		return
	case "DELETE":
		delete(s.modules, moduleID)
		s.websockets.SendModuleDeleted(moduleID)
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}

func (s *server) devicesHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	switch r.Method {
	case "GET":
		s.mu.Lock()
		defer s.mu.Unlock()

		getResp(w, r, s.devices, true)
		return
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}

func (s *server) deviceHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]

	s.mu.Lock()
	defer s.mu.Unlock()
	device, exists := s.devices[deviceID]
	switch r.Method {
	case "GET":
		getResp(w, r, device, exists)
		return
	case "PUT":
		updatedDevice := &v1alpha1.Device{}
		applyPut(w, r, updatedDevice, func() error {
			if exists && device.Name != updatedDevice.Name {
				http.Error(w, "Name in request does not match Name in path", http.StatusBadRequest)
			}
			if exists {
				s.devices[deviceID].Spec = updatedDevice.Spec
			} else {
				s.devices[deviceID] = updatedDevice
			}

			if moduleName, ok := s.deviceModules[deviceID]; ok {
				if !exists {
					s.websockets.SendModuleUpdated(moduleName)
				}
				for _, input := range updatedDevice.Spec.Inputs {
					s.websockets.SendValueChanged(moduleName + "." + deviceID + "." + input.Name, input.Value)
				}
			}
			return nil
		})
	case "DELETE":
		delete(s.devices, deviceID)
		w.WriteHeader(http.StatusOK)
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
	device, exists := s.devices[deviceID]
	if !exists {
		http.NotFound(w, r)
		return
	}

	input, exists := getValue(device.Status.ObservedInputs, inputID)
	if !exists {
		http.NotFound(w, r)
		return
	}
	
	switch r.Method {
	case "GET":
		getResp(w, r, input, true)
		return
	case "PUT":
		valueIn := &v1alpha1.Value{}
		applyPut(w, r, valueIn, func() error {
			if !setValue(device.Status.ObservedInputs, inputID, valueIn) {
				if t, ok := typesMap[inputID]; ok {
					device.Status.ObservedInputs = append(device.Status.ObservedInputs, v1alpha1.Value{Name: inputID, Type: t, Value: valueIn.Value})
				}
			}
			if moduleName, ok := s.deviceModules[deviceID]; ok {
				s.websockets.SendValueChanged(moduleName + "." + deviceID + "." + inputID, valueIn.Value)
			}
			return nil
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
	deviceID := vars["deviceID"]
	outputID := vars["outputID"]

	s.mu.Lock()
	defer s.mu.Unlock()
	device, exists := s.devices[deviceID]
	if !exists {
		http.NotFound(w, r)
		return
	}

	output, exists := getValue(device.Status.Outputs, outputID)
	if !exists {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case "GET":
		getResp(w, r, output, true)
		return
	case "PUT":
		valueIn := &v1alpha1.Value{}
		applyPut(w, r, valueIn, func() error {
			//log.Printf("putting output %v %v %v", deviceID, outputID, valueIn)
			setValue(device.Status.Outputs, outputID, valueIn)
			if moduleName, ok := s.deviceModules[deviceID]; ok {
				s.websockets.SendValueChanged(moduleName + "." + deviceID + "." + outputID, valueIn.Value)
			}
			return nil
		})
		return
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
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

func getResp(w http.ResponseWriter, r *http.Request, obj interface{}, exists bool) {
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

func getValue(list []v1alpha1.Value, name string) (v1alpha1.Value, bool) {
	for _, v := range list {
		if v.Name == name {
			return v, true
		}
	}
	return v1alpha1.Value{}, false
}

func setValue(list []v1alpha1.Value, name string, value *v1alpha1.Value) bool {
	for i, v := range list {
		if v.Name == name {
			list[i].Value = value.Value
			return true
		}
	}
	return false
}

func quantity(s string) resource.Quantity {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		log.Fatalf("failed to parse quantity: %s", s)
	}
	return q
}

var typesMap = map[string]v1alpha1.Type{
	"alarm":       v1alpha1.BooleanType,
	"activeCount": v1alpha1.IntegerType,
	"pressure":    v1alpha1.FloatType,
}