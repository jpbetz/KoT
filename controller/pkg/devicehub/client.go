/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package devicehub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jpbetz/KoT/controller/pkg/apis/k8softhings/v1alpha1"
)

type Client struct {
	Url    string
	Client *http.Client
}

func (c *Client) CheckDeviceStatuses() (*DeviceHub, error) {
	req, err := http.NewRequest(http.MethodGet, c.Url+ "/api/", nil)
	if err != nil {
		return nil, err
	}
	response, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("non-200 response code: %d", response.StatusCode)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	devices := &DeviceHub{}
	err = json.Unmarshal(body, devices)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (c *Client) CheckDeviceStatus(deviceID string) (v1alpha1.DeviceStatus, error) {
	req, err := http.NewRequest(http.MethodGet, c.Url+ "/api/", nil)
	if err != nil {
		return v1alpha1.DeviceStatus{}, err
	}
	response, err := c.Client.Do(req)
	if err != nil {
		return v1alpha1.DeviceStatus{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return v1alpha1.DeviceStatus{}, fmt.Errorf("non-200 response code: %d", response.StatusCode)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return v1alpha1.DeviceStatus{}, err
	}
	devices := &DeviceHub{}
	err = json.Unmarshal(body, devices)
	if err != nil {
		return v1alpha1.DeviceStatus{}, err
	}
	for _, device := range devices.Devices {
		if device.ID == deviceID {
			return ToDeviceStatus(device), nil
		}
	}
	return v1alpha1.DeviceStatus{}, fmt.Errorf("device not found, ID: %s (%#+v)", deviceID, devices)
}

func ToDeviceStatus(device *Device) v1alpha1.DeviceStatus {
	status := v1alpha1.DeviceStatus{}
	for _, input := range device.Inputs {
		status.Inputs = append(status.Inputs, &v1alpha1.Input{Name: input.ID, Value: input.Value})
	}
	for _, output := range device.Outputs {
		status.Outputs = append(status.Outputs, &v1alpha1.Output{Name: output.ID, Value: output.Value})
	}
	return status
}

func (c *Client) SetDeviceInput(deviceID string, inputID string, value float64) error {
	data, err := json.Marshal(&Input{Value: value})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/devices/%s/inputs/%s", c.Url, deviceID, inputID), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	response, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode < 200 && response.StatusCode >= 300 {
		return fmt.Errorf("non-200 response code: %d", response.StatusCode)
	}
	return nil
}

type DeviceHub struct {
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