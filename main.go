package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var config Config
var configFile = "config.json"
var (
	logger    *log.Logger
	logTarget string // "console" or "file"
	logFile   *os.File
)

type Config struct {
	ClientID          string `json:"clientId"`
	ClientSecret      string `json:"clientSecret"`
	APIUrlBase        string `json:"apiUrlBase"`
	AccessToken       string `json:"accessToken"`
	RefreshToken      string `json:"refreshToken"`
	ExpiresAt         int64  `json:"expiresAt"`
	WebhookToken      string `json:"webhookToken"`
	WeatherApiUrlBase string `json:"weatherApiUrlBase"`
	LogTarget         string `json:"logTarget"`
	LogFile           string `json:"logFile"`
	ServerPort        string `json:"serverPort"`
}

type WebhookCallback struct {
	ObjectType string `json:"object_type"`
	ObjectID   int    `json:"object_id"`
	AspectType string `json:"aspect_type"`
	OwnerID    int    `json:"owner_id"`
}

type ActivityResponse struct {
	Name           string    `json:"name"`
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

func ReadConfig(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	return nil
}

func SaveConfig(filename string) error {
	updatedJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, updatedJSON, 0644)
}

func RefreshToken(filename string) error {
	// Check if the token is expired
	currentTime := time.Now().Unix()
	if config.ExpiresAt > currentTime {
		expirationTime := time.Unix(config.ExpiresAt, 0).Format(time.RFC3339)
		logMessage("token is still valid. Expires at: %s", expirationTime)
		return nil
	}
	logMessage("refresh token.")

	url := config.APIUrlBase + "oauth/token"

	// Request body
	data := map[string]string{
		"client_id":     config.ClientID,
		"client_secret": config.ClientSecret,
		"grant_type":    "refresh_token",
		"refresh_token": config.RefreshToken,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// HTTP request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Parse response
	var tokenResponse TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return err
	}

	// Update config with new tokens
	config.AccessToken = tokenResponse.AccessToken
	config.RefreshToken = tokenResponse.RefreshToken
	config.ExpiresAt = tokenResponse.ExpiresAt

	// Save updated config
	err = SaveConfig(filename)
	if err != nil {
		return err
	}

	logMessage("token successfully refreshed.")
	return nil
}

func SendActivityUpdate(activityID string, activity ActivityResponse, emoji string, temp string) error {
	// Access-Token nur erneuern, wenn nötig
	err := RefreshToken(configFile)
	if err != nil {
		return fmt.Errorf("Fehler beim Aktualisieren des Tokens: %v", err)
	}

	apiURL := config.APIUrlBase + "activities/" + activityID
	newName := activity.Name + " " + emoji
	newDescription := activity.Description

	if temp != "999" {
		newDescription = activity.Description + fmt.Sprintf(" 🌡️ %s°C", temp)
	}

	logMessage("Send activity %s new name: %s, new description: %s", activityID, newName, newDescription)

	// Request-Daten erstellen
	updateRequest := ActivityResponse{
		Name:        newName,
		Description: newDescription,
	}

	// Request-Body in JSON umwandeln
	requestBody, err := json.Marshal(updateRequest)
	if err != nil {
		return fmt.Errorf("Fehler beim Marshaling der Anfrage: %v", err)
	}

	// Neue PUT-Anfrage erstellen
	req, err := http.NewRequest("PUT", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("Fehler beim Erstellen der Anfrage: %v", err)
	}

	// Header setzen
	req.Header.Set("Authorization", "Bearer "+config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Anfrage ausführen
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Fehler beim Ausführen des Requests: %v", err)
	}
	defer resp.Body.Close()

	// HTTP-Statuscode prüfen
	if resp.StatusCode != http.StatusOK {
		// Versuchen, die Fehlermeldung aus dem Body zu lesen
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("Fehler: HTTP-Status %d (Antwort konnte nicht gelesen werden)", resp.StatusCode)
		}
		return fmt.Errorf("Fehler: HTTP-Status %d, Antwort: %s", resp.StatusCode, string(bodyBytes))
	}

	logMessage("Activity %s successfully updated: %s", activityID, newName)
	return nil
}

func fetchActivityData(activityID string) (ActivityResponse, error) {
	err := RefreshToken(configFile)
	if err != nil {
		return ActivityResponse{}, fmt.Errorf("Error refreshing token")
	}

	apiURL := config.APIUrlBase + "activities/" + activityID

	// HTTP-GET-Anfrage ausführen
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return ActivityResponse{}, fmt.Errorf("Error creating GET Request")
	}

	req.Header.Set("Authorization", "Bearer "+config.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ActivityResponse{}, fmt.Errorf("Error GET Actitvity: %v", err)
	}
	defer resp.Body.Close()

	// Prüfen, ob der Statuscode OK ist (200)
	if resp.StatusCode != http.StatusOK {
		return ActivityResponse{}, fmt.Errorf("Error unexpected HTTP Statuscode: %d", resp.StatusCode)
	}

	// Antwort auslesen
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ActivityResponse{}, fmt.Errorf("Error reading body")
	}

	// Das JSON in eine Struktur umwandeln
	var activity ActivityResponse
	err = json.Unmarshal(body, &activity)
	if err != nil {
		return ActivityResponse{}, fmt.Errorf("Error parsing response %v", err)
	}

	logMessage("GET strava activity successful: %s", apiURL)

	return activity, nil
}

func getWeatherEmojiAndTemp(activity ActivityResponse, date string, hour int) (string, string, error) {

	url := fmt.Sprintf("%s?latitude=%f&longitude=%f&hourly=weather_code,temperature_2m&start_date=%s&end_date=%s", config.WeatherApiUrlBase, activity.StartLatLon[0], activity.StartLatLon[1], date, date)

	// HTTP-Request ausführen
	resp, err := http.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("error GET weather data, url %s, %v", url, err)
	}
	defer resp.Body.Close()

	// HTTP-Statuscode prüfen
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("error GET weather data: http-statuscode: %d, url: %s", resp.StatusCode, url)
	}

	// API-Antwort einlesen
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("error reading weather api response: %v", err)
	}

	// JSON-Daten in die Struktur parsen
	var weatherData WeatherResponse
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		return "", "", fmt.Errorf("error parsing weather api response, url %s, %v", url, err)
	}

	weatherCode := 100 // default weather code
	var temp float32 = 999

	// ---------------- weather code ----------------
	// check if hour is in weatherCode array
	if hour < 0 || hour >= len(weatherData.Hourly.WeatherCode) {
		logMessage("No weather code for given date, using default - targetHour: %d, date: %s, url: %s", hour, date, url)

	} else {
		weatherCode = weatherData.Hourly.WeatherCode[hour]
	}

	// gather emoji for given weather code
	emoji, exists := WeatherMap[weatherCode]
	if !exists {
		logMessage("No emoji found for weatherCode: %d, url: %s, using default emoji", weatherCode, url)
		emoji = WeatherMap[100]
	}

	// ---------------- weather code ----------------
	// check if temp is in weatherCode array
	if hour < 0 || hour >= len(weatherData.Hourly.Temperature) {
		logMessage("No temperature for given date - targetHour: %d, date: %s, url: %s", hour, date, url)
	} else {
		temp = weatherData.Hourly.Temperature[hour]
	}

	temperature := fmt.Sprintf("%d", int(temp))

	if weatherCode != 100 {
		logMessage("GET weather successful, temp: %s, hour: %d, date: %s, url: %s", temperature, hour, date, url)
	}
	return emoji, temperature, nil
}

func transformDateTime(activity ActivityResponse) (string, int, error) {

	t, err := time.Parse("2006-01-02T15:04:05Z", activity.StartDateLocal)
	if err != nil {
		return "", 0, fmt.Errorf("error parsing date: %s, %v", activity.StartDateLocal, err)
	}

	date := t.Format("2006-01-02")
	fullHourStr := t.Format("15")

	// add half of elapsedTime
	fullHour, err := strconv.Atoi(fullHourStr)
	if err != nil {
		return "", 0, fmt.Errorf("error converting string fullHourStr: %s, %v", fullHourStr, err)
	}

	// ergebnis wird abgeschnitten, nicht gerundet
	fullHour = fullHour + (activity.ElapsedTime / 3600)

	logMessage("Using date: %s, calculated fullHour: %d", activity.StartDateLocal, fullHour)

	// Rückgabe des Datums und der vollen Stunde
	return date, fullHour, nil
}

func updateActivity(activityID string) {
	// getActivityMetaData
	activity, err := fetchActivityData(activityID)
	if err != nil {
		logMessage("Error fetchActivityData: %v", err)
		return
	}

	// transform activity data
	date, targetHour, err := transformDateTime(activity)
	if err != nil {
		logMessage("Error transformPolylineToLatLong: %v", err)
		return
	}

	// get weather emoji based on activity date, hour
	emoji, temp, err := getWeatherEmojiAndTemp(activity, date, targetHour)
	if err != nil {
		logMessage("Error getWeatherEmojiAndTemp %v", err)
		return
	}
	// send activity update to strava
	err = SendActivityUpdate(activityID, activity, emoji, temp)
	if err != nil {
		logMessage("Error getWeatherEmoji %v", err)
		return
	}
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		// Strava sendet eine GET-Anfrage zur Verifizierung
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")

		if mode == "subscribe" && token == config.WebhookToken {
			logMessage("✅ WEBHOOK_VERIFIED")
			response := map[string]string{"hub.challenge": challenge}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			logMessage("❌ Verifizierung fehlgeschlagen")
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
		return
	}

	if r.Method == http.MethodPost {
		var callback WebhookCallback
		err := json.NewDecoder(r.Body).Decode(&callback)
		if err != nil {
			http.Error(w, "Fehler beim Parsen des JSON", http.StatusBadRequest)
			return
		}

		logMessage("-------------------------------------------------------------")
		logMessage("received new webhook: %+v", callback)

		// Falls es sich um eine neue Aktivität handelt, rufe fetchActivityData auf
		if callback.AspectType == "create" && callback.ObjectType == "activity" {
			activityID := strconv.Itoa(callback.ObjectID)
			go updateActivity(activityID) // Asynchron ausführen, blockiert nicht den Webhook
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// Unbekannte Methoden ablehnen
	http.Error(w, "Error", http.StatusMethodNotAllowed)
}

func logMessage(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...) // Nachricht formatieren

	switch config.LogTarget {
	case "console":
		fmt.Println(message)
	case "file":
		logger.Println(message)
	default:
		fmt.Println(message)
	}
}

func init() {
	err := ReadConfig(configFile)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		os.Exit(1)
	}

	// log.Printf("Successful read config: %s.", configFile)

	// Log-Datei öffnen oder erstellen
	logFile, err = os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Error opening log file:  %s,%v", config.LogFile, err)
		os.Exit(1)
	}

	// Logger initialisieren
	logger = log.New(logFile, "", log.LstdFlags)

	// log.Println("Start Application")
}

func main() {
	defer logFile.Close() // auto close file after end of program

	http.HandleFunc("/webhook", WebhookHandler) // `/webhook` für GET (Verifikation) und POST (Events)
	logMessage("-------------------------------------------------------------")
	logMessage("server running on port %s...", config.ServerPort)
	log.Fatal(http.ListenAndServe(":"+config.ServerPort, nil))

}

/* Strukturierung des Codes
- Verwende sinnvolle Pakete:
		Teile deine Anwendung in verschiedene Pakete auf, um den Code modular und wartbar zu halten
- Vermeide lange Funktionen

myapp/
├── cmd/          # Enthält die Hauptanwendung
├── internal/     # Enthält private Pakete, die nicht exportiert werden sollen
├── pkg/          # Öffentliche, wiederverwendbare Pakete
├── api/          # Definition der API, z.B. für HTTP-Handler
├── model/        # Datenstrukturen und Geschäftslogik
├── service/      # Services für Geschäftslogik
├── config/       # Konfigurationsdateien und -logik
└── main.go       # Einstiegspunkt der Anwendung

*/

/* TEST
curl -X POST "http://localhost:8080/webhook" \
  -H "Content-Type: application/json" \
  -d '{
    "object_type": "activity",
    "object_id": 13579796478,
    "aspect_type": "create",
    "owner_id": 11111
  }'
*/
