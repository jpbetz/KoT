package types

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

type ValueChangedMessage struct {
	Path string `json:"path"`
	Value resource.Quantity `json:"value"`
}
