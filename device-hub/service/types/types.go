package types

import (
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/jpbetz/KoT/apis/things/v1alpha1"
)

type Modules struct {
	Modules []*Module `json:"modules"`
}

type Module struct {
	ID string `json:"id"`
	Pump *Device `json:"pump"`
	WaterAlarm *Device `json:"waterAlarm"`
	PressureSensor *Device `json:"pressureSensor"`
}

type Device struct {
	ID string `json:"id"`
	Inputs []v1alpha1.Value `json:"inputs"`
	Outputs []v1alpha1.Value `json:"outputs"`
}

type ValueChangedMessage struct {
	Path string `json:"path"`
	Value resource.Quantity `json:"value"`
}

func (m *Modules) GetDevice(id string) (*Module, *Device) {
	for _, module := range m.Modules {
		for _, device := range module.GetDevices() {
			if device.ID == id {
				return module, device
			}
		}
	}
	return nil, nil
}

func (m *Modules) GetModule(id string) *Module {
	for _, module := range m.Modules {
		if module.ID == id {
			return module
		}
	}
	return nil
}

func (m *Modules) PutModule(in *Module) {
	for i, module := range m.Modules {
		if module.ID == in.ID {
			m.Modules[i] = in
			return
		}
	}
	m.Modules = append(m.Modules, in)
}

func (m *Modules) DeleteModule(moduleID string) {
	for i, module := range m.Modules {
		if module.ID == moduleID {
			m.Modules = append(m.Modules[:i], m.Modules[i+1:]...)
			return
		}
	}
}

func (m *Module) GetDevices() []*Device {
	return []*Device{m.WaterAlarm, m.PressureSensor, m.Pump}
}

func (d *Device) GetInput(name string) (v1alpha1.Value, bool) {
	for _, input := range d.Inputs {
		if input.Name == name {
			return input, true
		}
	}
	return v1alpha1.Value{}, false
}

func (d *Device) SetInput(name string, q resource.Quantity) bool {
	for i, input := range d.Inputs {
		if input.Name == name {
			d.Inputs[i] = v1alpha1.Value{Name: name, Type: input.Type, Value: q}
			return true
		}
	}
	return false
}

func (d *Device) GetOutput(name string) (v1alpha1.Value, bool) {
	for _, output := range d.Outputs {
		if output.Name == name {
			return output, true
		}
	}
	return v1alpha1.Value{}, false
}

func (d *Device) SetOutput(name string, q resource.Quantity) bool {
	for i, output := range d.Outputs {
		if output.Name == name {
			d.Outputs[i] = v1alpha1.Value{Name: name, Type: output.Type, Value: q}
			return true
		}
	}
	return false
}