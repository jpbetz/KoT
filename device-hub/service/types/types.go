package types

type SimulatedDevices struct {
	Devices []*Device `json:"devices"`
}

type Device struct {
	ID string
	Inputs []*Input `json:"inputs"`
	Outputs []*Output `json:"outputs"`
}

type Input struct {
	ID string
	Value float64 `json:"value"`
}

type Output struct {
	ID string
	Value float64 `json:"value"`
}

type ValueChangedMessage struct {
	Path string `json:"path"`
	Value float64 `json:"value"`
}


func (d *SimulatedDevices) GetDevice(id string) *Device {
	for _, device := range d.Devices {
		if device.ID == id {
			return device
		}
	}
	return nil
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