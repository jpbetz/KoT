package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

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

	Items []Module `json:"items"`
}
