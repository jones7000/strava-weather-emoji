package model

type WebhookCallback struct {
	ObjectType string  `json:"object_type"`
	ObjectID   int     `json:"object_id"`
	AspectType string  `json:"aspect_type"`
	OwnerID    float32 `json:"owner_id"`
}

type ActivityResponse struct {
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Map            Map       `json:"map"`
	StartDateLocal string    `json:"start_date_local"` //"start_date_local": "2025-02-03T16:56:12Z",
	StartLatLon    []float32 `json:"start_latlng"`
	ElapsedTime    int       `json:"elapsed_time"`
	Description    string    `json:"description"`
}

type Map struct {
	ID              string `json:"id"`
	Polyline        string `json:"polyline"`
	ResourceState   int    `json:"resource_state"`
	SummaryPolyline string `json:"summary_polyline"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type WeatherResponse struct {
	Hourly struct {
		Time        []string  `json:"time"`
		WeatherCode []int     `json:"weather_code"`
		Temperature []float32 `json:"temperature_2m"`
	} `json:"hourly"`
}

var WeatherMap = map[int]string{ // https://open-meteo.com/ > weather codes
	0:   "☀️",     // Clear sky
	1:   "🌤",      // Mainly clear
	2:   "⛅",      // Partly cloudy
	3:   "☁️",     // Overcast
	45:  "🌫",      // Fog
	48:  "🌫❄️",    // Depositing rime fog
	51:  "🌦",      // Drizzle: Light
	53:  "🌧",      // Drizzle: Moderate
	55:  "🌧🌧",     // Drizzle: Dense
	56:  "🧊🌧",     // Freezing Drizzle: Light
	57:  "🧊🌧🌧",    // Freezing Drizzle: Dense
	61:  "🌦",      // Rain: Slight
	63:  "🌧",      // Rain: Moderate
	65:  "🌧🌧",     // Rain: Heavy
	66:  "🧊🌧",     // Freezing Rain: Light
	67:  "🧊🌧🌧",    // Freezing Rain: Heavy
	71:  "❄️",     // Snow fall: Slight
	73:  "❄️❄️",   // Snow fall: Moderate
	75:  "❄️❄️❄️", // Snow fall: Heavy
	77:  "🌨",      // Snow grains
	80:  "🌦",      // Rain showers: Slight
	81:  "🌧",      // Rain showers: Moderate
	82:  "🌧🌧",     // Rain showers: Violent
	85:  "🌨",      // Snow showers: Slight
	86:  "🌨🌨",     // Snow showers: Heavy
	95:  "⛈",      // Thunderstorm: Slight or moderate
	96:  "⛈🌨",     // Thunderstorm with slight hail
	99:  "⛈🌨🌨",    // Thunderstorm with heavy hail
	100: "🏃",      // unknown
}

var IndoorActivities = []string{
	"Crossfit", "Elliptical", "StairStepper", "VirtualRide",
	"VirtualRun", "WeightTraining", "Workout", "Yoga",
}

var RunningActivities = []string{
	"Run", "TrailRun", "Hike",
}
