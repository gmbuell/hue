package hue

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path"
)

type Bridge struct {
	host    string
	user    string
	client  http.Client
	baseUrl url.URL
}

func NewBridge(host string, user string) *Bridge {
	return &Bridge{host: host, user: user, baseUrl: url.URL{Scheme: "http", Host: host, Path: path.Join("api", user)}}
}

func (bridge *Bridge) GetAllLights() (Lights, error) {
	lightsUrl := bridge.baseUrl
	lightsUrl.Path = path.Join(lightsUrl.Path, "lights")
	resp, err := bridge.client.Get(lightsUrl.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Unmarshal
	var lights Lights
	err = json.NewDecoder(resp.Body).Decode(&lights)
	if err != nil {
		return nil, err
	}
	return lights, nil
}

func (bridge *Bridge) GetLight(id string) (*Light, error) {
	lightsUrl := bridge.baseUrl
	lightsUrl.Path = path.Join(lightsUrl.Path, "lights", id)
	resp, err := bridge.client.Get(lightsUrl.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Unmarshal
	light := Light{bridge: bridge, index: id}
	err = json.NewDecoder(resp.Body).Decode(&light)
	if err != nil {
		return nil, err
	}
	return &light, nil
}

func (bridge *Bridge) SetLightName(id string, name string) error {
	lightsUrl := bridge.baseUrl
	lightsUrl.Path = path.Join(lightsUrl.Path, "lights", id)
	var putBody bytes.Buffer
	err := json.NewEncoder(&putBody).Encode(map[string]string{"name": name})
	if err != nil {
		return err
	}

	setRequest, err := http.NewRequest("PUT", lightsUrl.String(), &putBody)
	if err != nil {
		return err
	}

	resp, err := bridge.client.Do(setRequest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Unmarshal
	var response []map[string]map[string]string
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	if len(response) != 1 {
		return errors.New("Expected one result in SetLightName response.")
	}

	if _, ok := response[0]["success"]; !ok {
		return errors.New("API did not return success")
	}

	return nil
}
