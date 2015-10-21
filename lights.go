package hue

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"path"
)

const (
	COLORMODE_CT = "ct"
	COLORMODE_HS = "hs"
	COLORMODE_XY = "xy"
)

const (
	ALERT_NONE    = "none"
	ALERT_SELECT  = "select"
	ALERT_LSELECT = "lselect"
)

const (
	EFFECT_NONE      = "none"
	EFFECT_COLORLOOP = "colorloop"
)

type Light struct {
	ModelId     string   `json:"modelid"`
	Name        string   `json:"name"`
	PointSymbol struct{} `json:"pointsymbol"`
	State       struct {
		Alert            string    `json:"alert"`
		Brightness       uint8     `json:"bri"`
		ColorMode        string    `json:"colormode"`
		ColorTemperature uint16    `json:"ct"`
		Effect           string    `json:"effect"`
		Hue              uint16    `json:"hue"`
		On               bool      `json:"on"`
		Reachable        bool      `json:"reachable"`
		Saturation       uint8     `json:"sat"`
		XY               []float64 `json:"xy"`
	} `json:"state"`
	SoftwareVersion   string  `json:"swversion"`
	Type              string  `json:"type"`
	UniqueId          string  `json:"uniqueid"`
	ManufacturerName  string  `json:"manufacturername"`
	LuminaireUniqueId string  `json:"luminaireuniqueid"`
	Bridge            *Bridge `json:"-"`
	Index             string  `json:"-"`
}

func (light *Light) SetName(name string) (map[string]interface{}, error) {
	lightsUrl := light.Bridge.baseUrl
	lightsUrl.Path = path.Join(lightsUrl.Path, "lights", light.Index)
	var putBody bytes.Buffer
	err := json.NewEncoder(&putBody).Encode(map[string]string{"name": name})
	if err != nil {
		return nil, err
	}

	setRequest, err := http.NewRequest("PUT", lightsUrl.String(), &putBody)
	if err != nil {
		return nil, err
	}

	resp, err := light.Bridge.client.Do(setRequest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Unmarshal
	var response []APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if len(response) != 1 {
		return nil, errors.New("Expected one result in SetLightName response.")
	}

	if response[0].Error != nil {
		return nil, response[0].Error
	}

	return response[0].Result, nil
}

func (light *Light) SetState(options ...func(*StateConfig) error) (map[string]interface{}, error) {
	stateUrl := light.Bridge.baseUrl
	stateUrl.Path = path.Join(stateUrl.Path, "lights", light.Index, "state")
	var putBody bytes.Buffer
	state := make(StateConfig)
	for _, option := range options {
		err := option(&state)
		if err != nil {
			return nil, err
		}
	}

	// Make sure we are actually setting something relevant.
	if len(state) == 0 {
		return nil, nil
	} else if len(state) == 1 {
		if _, ok := state["transitiontime"]; ok {
			// We are only setting the transition.
			return nil, nil
		}
	} else {
		// Make sure we do not try and turn lights on/off concurrently with other changes.
		if _, ok := state["on"]; ok {
			return nil, errors.New("Cannot turn light on/off concurrently with other changes.")
		}
	}

	err := json.NewEncoder(&putBody).Encode(state)
	if err != nil {
		return nil, err
	}

	setRequest, err := http.NewRequest("PUT", stateUrl.String(), &putBody)
	if err != nil {
		return nil, err
	}

	resp, err := light.Bridge.client.Do(setRequest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Unmarshal
	var response []APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, errors.New("API did not return success.")
	}

	updateValues := make(map[string]interface{})
	for _, responseItem := range response {
		if responseItem.Error != nil {
			return nil, responseItem.Error
		}
		for updatePath, updateValue := range responseItem.Result {
			updateValues[updatePath] = updateValue
		}
	}

	return updateValues, nil
}
