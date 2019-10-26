package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Device is a specification for a Device resource
type Device struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceSpec   `json:"spec"`
	Status DeviceStatus `json:"status"`
}

// DeviceSpec is the spec for a Sensor resource
type DeviceSpec struct {
	Inputs []*Input `json:"inputs,omitempty"`
}

// DeviceStatus is the status for a Sensor resource
type DeviceStatus struct {
	Inputs []*Input `json:"inputs,omitempty"`
	Outputs []*Output `json:"outputs,omitempty"`
}

type Input struct {
	Name string `json:"name"`
	Value float64 `json:"value"`
}

type Output struct {
	Name string `json:"name"`
	Value float64 `json:"value"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceList is a list of Sensor resources
type DeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Device `json:"items"`
}