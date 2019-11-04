package types

type EventMessage struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
	Value string `json:"value,omitempty"`
}