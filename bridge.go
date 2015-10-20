package hue

import (
	"encoding/json"
	"errors"
	"fmt"
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

type BridgeResponse struct {
	Id         string `json:"id"`
	InternalIP string `json:"internalipaddress"`
	MACAddress string `json:"macaddress"`
	Name       string `json:"name"`
}

func NewBridge(host string, user string) *Bridge {
	return &Bridge{host: host, user: user, baseUrl: url.URL{Scheme: "http", Host: host, Path: path.Join("api", user)}}
}

func BridgeFromNUPnP(user string) (*Bridge, error) {
	var client http.Client
	resp, err := client.Get("https://www.meethue.com/api/nupnp")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var bridgeResponses []BridgeResponse
	err = json.NewDecoder(resp.Body).Decode(&bridgeResponses)
	if err != nil {
		return nil, err
	}

	// Use the first bridge
	if len(bridgeResponses) > 0 {
		return NewBridge(bridgeResponses[0].InternalIP, user), nil
	} else {
		return nil, errors.New("No bridges returned from N-UPnP.")
	}
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

type APIResponse struct {
	Result map[string]interface{} `json:"success,omitempty"`
	Error  *ErrorResponse         `json:"error,omitempty"`
}

type ErrorResponse struct {
	Address     string `json:"address"`
	Description string `json:"description"`
	Type        int16  `json:"type"`
}

func (err ErrorResponse) Error() string {
	return fmt.Sprintf("API did not return success (%s): %s)", err.Address, err.Description)
}
