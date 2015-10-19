package hue

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
	SoftwareVersion   string `json:"swversion"`
	Type              string `json:"type"`
	UniqueId          string `json:"uniqueid"`
	ManufacturerName  string `json:"manufacturername"`
	LuminaireUniqueId string `json:"luminaireuniqueid"`
}