package hue

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

type Lights map[string]Light

type Light struct {
	ModelId     string `json:"modelid"`
	Name        string `json:"name"`
	PointSymbol struct {
		_ string `json:"1"`
		_ string `json:"2"`
		_ string `json:"3"`
		_ string `json:"4"`
		_ string `json:"5"`
		_ string `json:"6"`
		_ string `json:"7"`
		_ string `json:"8"`
	} `json:"pointsymbol"`
	State struct {
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
	bridge            *Bridge `json:"-"`
	index             string  `json:"-"`
}

func (light *Light) SetIndex(index string) error {
	if len(light.index) > 0 {
		return errors.New("Light already has an index.")
	}
	light.index = index
	return nil
}

type SetState struct {
	Alert                 string    `json:"alert,omitempty"`
	Brightness            uint8     `json:"bri,omitempty"`
	ColorMode             string    `json:"colormode,omitempty"`
	ColorTemperature      uint16    `json:"ct,omitempty"`
	Effect                string    `json:"effect,omitempty"`
	Hue                   uint16    `json:"hue,omitempty"`
	On                    bool      `json:"on"`
	Reachable             bool      `json:"reachable,omitempty"`
	Saturation            uint8     `json:"sat,omitempty"`
	XY                    []float64 `json:"xy,omitempty"`
	TransitionTime        uint16    `json:"transitiontime,omitempty"`
	BrightnessDelta       int16     `json:"bri_inc,omitempty"`
	SaturationDelta       int16     `json:"sat_inc,omitempty"`
	HueDelta              int16     `json:"hue_inc,omitempty"`
	ColorTemperatureDelta int16     `json:"ct_inc,omitempty"`
	XYDelta               []float64 `json:"xy_inc,omitempty"`
}

func (light *Light) SetName(name string) error {
	lightsUrl := light.bridge.baseUrl
	lightsUrl.Path = path.Join(lightsUrl.Path, "lights", light.index)
	var putBody bytes.Buffer
	err := json.NewEncoder(&putBody).Encode(map[string]string{"name": name})
	if err != nil {
		return err
	}

	setRequest, err := http.NewRequest("PUT", lightsUrl.String(), &putBody)
	if err != nil {
		return err
	}

	resp, err := light.bridge.client.Do(setRequest)
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
		return errors.New("API did not return success.")
	}

	return nil
}

func (light *Light) SetState(state SetState) (map[string]interface{}, error) {
	stateUrl := light.bridge.baseUrl
	stateUrl.Path = path.Join(stateUrl.Path, "lights", light.index, "state")
	var putBody bytes.Buffer
	err := json.NewEncoder(&putBody).Encode(state)
	if err != nil {
		return nil, err
	}

	setRequest, err := http.NewRequest("PUT", stateUrl.String(), &putBody)
	if err != nil {
		return nil, err
	}

	resp, err := light.bridge.client.Do(setRequest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Unmarshal
	var response []map[string]map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, errors.New("API did not return success.")
	}

	updateValues := make(map[string]interface{})
	for _, responseItem := range response {
		responseValueMap, isSuccess := responseItem["success"]
		if !isSuccess {
			return nil, errors.New(fmt.Sprintf("API did not return success: %+v", responseItem))
		}
		for updatePath, updateValue := range responseValueMap {
			updateValues[updatePath] = updateValue
		}
	}

	return updateValues, nil
}
