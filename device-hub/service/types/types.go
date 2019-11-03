package types

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
	Inputs []*Input `json:"inputs"`
	Outputs []*Output `json:"outputs"`
}

type Input struct {
	ID string `json:"id"`
	Value float64 `json:"value"`
}

type Output struct {
	ID string `json:"id"`
	Value float64 `json:"value"`
}

type ValueChangedMessage struct {
	Path string `json:"path"`
	Value float64 `json:"value"`
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

func (d *Device) GetInput(id string) *Input {
	for _, input := range d.Inputs {
		if input.ID == id {
			return input
		}
	}
	return nil
}

func (d *Device) GetOutput(id string) *Output {
	for _, output := range d.Outputs {
		if output.ID == id {
			return output
		}
	}
	return nil
}