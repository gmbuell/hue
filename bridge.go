package hue

import (
	"encoding/json"
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
	for key, val := range lights {
		err := val.SetIndex(key)
		if err != nil {
			return nil, err
		}
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

type ErrorResponse struct {
	Address     string `json:"address"`
	Description string `json:"description"`
	Type        int16  `json:"type"`
}
