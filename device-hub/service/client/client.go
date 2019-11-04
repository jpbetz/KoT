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

	deepseav1alpha1 "github.com/jpbetz/KoT/apis/deepsea/v1alpha1"
	"github.com/jpbetz/KoT/apis/things/v1alpha1"
)

type Client struct {
	Url    string
	Client *http.Client
}

func (c *Client) GetModule(name string) (*deepseav1alpha1.Module, error) {
	module := &deepseav1alpha1.Module{}
	found, err := c.get("/api/modules/" + name, module)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return module, nil
}

func (c *Client) PutModule(module *deepseav1alpha1.Module) error {
	return c.put("/api/modules/" + module.Name, module)
}

func (c *Client) GetDevice(name string) (*v1alpha1.Device, error) {
	device := &v1alpha1.Device{}
	found, err := c.get("/api/devices/" + name, device)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return device, nil
}

func (c *Client) PutDevice(device *v1alpha1.Device) error {
	return c.put("/api/devices/" + device.Name, device)
}

func (c *Client) put(path string, content interface{}) error {
	data, err := json.Marshal(content)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, c.Url + path, bytes.NewBuffer(data))
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

func (c *Client) get(path string, out interface{}) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, c.Url + path, nil)
	if err != nil {
		return false, err
	}
	response, err := c.Client.Do(req)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}
	if response.StatusCode == 404 {
		return false, nil
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return false, fmt.Errorf("non-200 response code: %d", response.StatusCode)
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		return false, err
	}
	return true, nil
}