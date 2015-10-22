package hue

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	GAMUT_A_RED   = [2]float64{0.704, 0.296}
	GAMUT_A_GREEN = [2]float64{0.2151, 0.7106}
	GAMUT_A_BLUE  = [2]float64{0.138, 0.08}
	GAMUT_B_RED   = [2]float64{0.675, 0.322}
	GAMUT_B_GREEN = [2]float64{0.409, 0.518}
	GAMUT_B_BLUE  = [2]float64{0.167, 0.04}
	GAMUT_C_RED   = [2]float64{0.675, 0.322}
	GAMUT_C_GREEN = [2]float64{0.2151, 0.7106}
	GAMUT_C_BLUE  = [2]float64{0.167, 0.04}
)

type StateConfig map[string]interface{}

func On(lightOn bool) func(*StateConfig) error {
	return func(statePtr *StateConfig) error {
		state := *statePtr
		state["on"] = lightOn
		if len(state) != 1 {
			return errors.New("Cannot turn light on/off concurrently with other changes.")
		}
		return nil
	}
}

func (light *Light) BrightnessDelta(delta int16) func(*StateConfig) error {
	return func(statePtr *StateConfig) error {
		state := *statePtr
		supportedVersion, err := version.NewVersion("1.2.1")
		if err != nil {
			return err
		}
		bridgeVersion, err := version.NewVersion(light.Bridge.config.APIVersion)
		if err != nil {
			return err
		}
		if bridgeVersion.LessThan(supportedVersion) {
			brightness := int16(light.State.Brightness) + delta
			if brightness > 254 {
				if light.State.Brightness == 254 {
					// Don't set any state if we are already at max brightness
					return nil
				}
				brightness = 254
			} else if brightness < 1 {
				if light.State.Brightness == 1 {
					// Don't set any state if we are already at min brightness
					return nil
				}
				brightness = 1
			}
			state["bri"] = uint8(brightness)
			light.State.Brightness = uint8(brightness)
		} else {
			state["bri_inc"] = delta
		}
		return nil
	}
}

func Brightness(brightness uint8) func(*StateConfig) error {
	return func(statePtr *StateConfig) error {
		state := *statePtr
		if brightness > 254 {
			return errors.New(fmt.Sprintf("Invalid brightness %d > 254", brightness))
		} else if brightness < 1 {
			return errors.New(fmt.Sprintf("Invalid brightness %d < 1", brightness))
		}
		state["bri"] = brightness
		return nil
	}
}

func TransitionTime(time uint16) func(*StateConfig) error {
	return func(statePtr *StateConfig) error {
		state := *statePtr
		state["transitiontime"] = time
		return nil
	}
}

func XY(xy [2]float64) func(*StateConfig) error {
	return func(statePtr *StateConfig) error {
		state := *statePtr
		state["xy"] = xy
		return nil
	}
}

func Saturation(saturation uint8) func(*StateConfig) error {
	return func(statePtr *StateConfig) error {
		state := *statePtr
		if saturation > 254 {
			return errors.New(fmt.Sprintf("Invalid saturation %d > 254", saturation))
		} else if saturation < 1 {
			return errors.New(fmt.Sprintf("Invalid saturation %d < 1", saturation))
		}
		state["sat"] = saturation
		return nil
	}
}

func Hue(hue uint16) func(*StateConfig) error {
	return func(statePtr *StateConfig) error {
		state := *statePtr
		state["hue"] = hue
		return nil
	}
}

func Color(color colorful.Color) func(*StateConfig) error {
	return func(statePtr *StateConfig) error {
		state := *statePtr
		x, y, z := color.Clamped().Xyz()
		hueX := x / (x + y + z)
		hueY := y / (x + y + z)
		state["xy"] = [2]float64{hueX, hueY}
		return nil
	}
}

func BrightnessFromColor(color colorful.Color) func(*StateConfig) error {
	return func(statePtr *StateConfig) error {
		state := *statePtr
		_, y, _ := color.Clamped().Xyz()
		brightness := y * 254
		if brightness > 254.0 {
			brightness = 254.0
		} else if brightness < 1.0 {
			brightness = 1.0
		}
		state["bri"] = uint8(brightness)
		return nil
	}
}
