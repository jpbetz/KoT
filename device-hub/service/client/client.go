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

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jpbetz/KoT/apis/things/v1alpha1"
	"github.com/jpbetz/KoT/device-hub/service/types"
)

type Client struct {
	Url    string
	Client *http.Client
}

func (c *Client) GetModule(name string) (*types.Module, error) {
	module := &types.Module{}
	err := c.get(c.Url+ "/api/modules/%s", nil)
	if err != nil {
		return nil, err
	}
	return module, nil
}

// TODO: remove module
func (c *Client) PutModule(module *v1alpha1.Module) error {
	return c.put(fmt.Sprintf("%s/api/modules/%s", c.Url, module.Name), module)
}

func (c *Client) CheckDeviceStatus(deviceID string) (v1alpha1.DeviceStatus, error) {
	req, err := http.NewRequest(http.MethodGet, c.Url+ "/api/modules", nil)
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
	modules := &types.Modules{}
	err = json.Unmarshal(body, modules)
	if err != nil {
		return v1alpha1.DeviceStatus{}, err
	}
	for _, module := range modules.Modules {
		for _, device := range []*types.Device{module.PressureSensor, module.WaterAlarm, module.Pump} {
			if device.ID == deviceID {
				return ToDeviceStatus(device), nil
			}
		}
	}
	return v1alpha1.DeviceStatus{}, fmt.Errorf("device not found, ID: %s (%#+v)", deviceID, modules)
}

func ToDeviceStatus(device *types.Device) v1alpha1.DeviceStatus {
	status := v1alpha1.DeviceStatus{}
	status.ObservedInputs = device.Inputs
	status.Outputs = device.Outputs
	return status
}

func (c *Client) SetDeviceInput(deviceID string, inputID string, value v1alpha1.Value) error {
	return c.put(fmt.Sprintf("%s/api/devices/%s/inputs/%s", c.Url, deviceID, inputID), value)
}

func (c *Client) put(url string, content interface{}) error {
	data, err := json.Marshal(content)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
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

func (c *Client) get(url string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("non-200 response code: %d", response.StatusCode)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		return err
	}
	return nil
}