package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Device is a specification for a device in a deep see station. It can be a
// sensor (if it only has outputs), or an actor (if it only has inputs), or both.
type Device struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceSpec   `json:"spec"`
	Status DeviceStatus `json:"status"`
}

// DeviceSpec is the spec for a Sensor resource
type DeviceSpec struct {
	// inputs are the desired value for an actor.
	Inputs []Value `json:"inputs,omitempty"`
}

// DeviceStatus is the status for a Sensor resource
type DeviceStatus struct {
	// observedInputs are the inputs the device observed.
	ObservedInputs []Value `json:"observedInputs,omitempty"`
	// outputs are values of a sensor.
	Outputs []Value `json:"outputs,omitempty"`
}

// Value is a named and typed value.
type Value struct {
	// name is the name of this input value.
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// value is the floating point input value.
	// +kubebuilder:validation:Required
	Value float64 `json:"value"`
	// +kubebuilder:validation:Default=float
	// +kubebuilder:validation:Enum={"integer","float","boolean"}
	Type string `json:"type"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceList is a list of Sensor resources
type DeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Device `json:"items"`
}

// Module is a section in the deep see station, connected to a number of devices.
type Module struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec   ModuleSpec   `json:"spec"`
	Status ModuleStatus `json:"status"`
}

// ModuleSpec defines the devices in the deep see station module.
type ModuleSpec struct {
	// devices specifies the
	// +kubebuilder:validation:Required
	Devices ModuleDevices `json:"devices"`
}

// ModuleDevices describes the devices this module is connected to.
type ModuleDevices struct {
	// pump is the pump device name. It is of type integer.
	// +kubebuilder:validation:Required
	Pump string `json:"pump"`
	// waterAlarm is the water alarm device name. It is of type boolean.
	// +kubebuilder:validation:Required
	WaterAlarm string `json:"waterAlarm"`
	// pressureSensor is the pressure sensor device name. It is of type float.
	// +kubebuilder:validation:Required
	PressureSensor string `json:"pressureSensor"`
}

// ModuleStatus is the status of the deep see station module.
type ModuleStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceList is a list of Sensor resources
type ModuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Device `json:"items"`
}
