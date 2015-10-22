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
	config  bridgeConfig
}

type bridgeConfig struct {
	Name     string `json:"name"`
	SWUpdate struct {
		CheckForUpdate bool `json:"checkforupdate"`
		DeviceTypes    struct {
			Bridge bool    `json:"bridge"`
			lights []Light `json:"lights"`
		} `json:"devicetypes"`
	} `json:"swupdate"`
	Whitelist        struct{} `json:"whitelist"`
	APIVersion       string   `json:"apiversion"`
	SWVersion        string   `json:"swversion"`
	ProxyAddress     string   `json:"proxyaddress"`
	ProxyPort        uint16   `json:"proxyport"`
	LinkButton       bool     `json:"linkbutton"`
	IPAddress        string   `json:"ipaddress"`
	MAC              string   `json:"mac"`
	Netmask          string   `json:"netmask"`
	Gateway          string   `json:"gateway"`
	DHCP             bool     `json:"dhcp"`
	PortalServices   bool     `json:"portalservices"`
	UTC              string   `json:"UTC"`
	LocalTime        string   `json:"localtime"`
	TimeZone         string   `json:"timezone"`
	ZigbeeChannel    uint8   `json:"zigbeechannel"`
	TouchLink        bool     `json:"touchlink"`
	FactoryNew       bool     `json:"factorynew"`
	ReplacesBridgeID string   `json:"replacesbridgeid"`
}

type BridgeResponse struct {
	Id         string `json:"id"`
	InternalIP string `json:"internalipaddress"`
	MACAddress string `json:"macaddress"`
	Name       string `json:"name"`
}

func NewBridge(host string, user string) (*Bridge, error) {
	baseUrl := url.URL{Scheme: "http", Host: host, Path: path.Join("api", user)}
	configURL := baseUrl
	configURL.Path = path.Join(configURL.Path, "config")
	resp, err := http.Get(configURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Unmarshal
	var config bridgeConfig
	err = json.NewDecoder(resp.Body).Decode(&config)
	if err != nil {
		return nil, err
	}
	if len(config.APIVersion) == 0 {
		config.APIVersion = "1.0"
	}
	fmt.Printf("Bridge API version: %s\n", config.APIVersion)
	return &Bridge{host: host, user: user, baseUrl: baseUrl, config: config}, nil
}

func BridgeFromNUPnP(user string) (*Bridge, error) {
	resp, err := http.Get("https://www.meethue.com/api/nupnp")
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
		return NewBridge(bridgeResponses[0].InternalIP, user)
	} else {
		return nil, errors.New("No bridges returned from N-UPnP.")
	}
}

func (bridge *Bridge) GetAllLights() ([]*Light, error) {
	lightsUrl := bridge.baseUrl
	lightsUrl.Path = path.Join(lightsUrl.Path, "lights")
	resp, err := bridge.client.Get(lightsUrl.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Unmarshal
	var lights map[string]*Light
	err = json.NewDecoder(resp.Body).Decode(&lights)
	if err != nil {
		return nil, err
	}
	var lightsArray []*Light
	for key, val := range lights {
		val.Index = key
		val.Bridge = bridge
		lightsArray = append(lightsArray, val)
	}
	return lightsArray, nil
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
	light := Light{Bridge: bridge, Index: id}
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
