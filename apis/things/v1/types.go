package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
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

// Value is named and typed value. Only one of float, boolean or integer can be set.
type Value struct {
	// name is the name of this input value.
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// float is a floating point input value.
	// +kubebuilder:validation:Type=number
	Float *resource.Quantity `json:"float"`
	// boolean is a true or false value.
	Boolean *bool `json:"boolean"`
	// integer is a integer value.
	Integer *int32 `json:"float"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceList is a list of Sensor resources
type DeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Device `json:"items"`
}
