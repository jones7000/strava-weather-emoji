package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strava-api/config"
	"strava-api/logger"
	"strava-api/model"
)

func GetWeatherEmojiAndTemp(activity model.ActivityResponse, date string, hour int) (string, string, error) {
	cfg, _ := config.GetConfig()
	url := fmt.Sprintf("%s?latitude=%f&longitude=%f&hourly=weather_code,temperature_2m&start_date=%s&end_date=%s", cfg.WeatherApiUrlBase, activity.StartLatLon[0], activity.StartLatLon[1], date, date)

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
	var weatherData model.WeatherResponse
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		return "", "", fmt.Errorf("error parsing weather api response, url %s, %v", url, err)
	}

	weatherCode := 100     // default weather code
	var temp float32 = 999 // default temp

	// ---------------- weather code ----------------
	// check if hour is in weatherCode array
	if hour < 0 || hour >= len(weatherData.Hourly.WeatherCode) {
		logger.LogMessage("No weather code for given date, using default - targetHour: %d, date: %s, url: %s", hour, date, url)

	} else {
		weatherCode = weatherData.Hourly.WeatherCode[hour]
	}

	// gather emoji for given weather code
	emoji, exists := model.WeatherMap[weatherCode]
	if !exists {
		logger.LogMessage("No emoji found for weatherCode: %d, url: %s, using default emoji", weatherCode, url)
		emoji = model.WeatherMap[100]
	}

	// ---------------- weather code ----------------
	// check if temp is in weatherCode array
	if hour < 0 || hour >= len(weatherData.Hourly.Temperature) {
		logger.LogMessage("No temperature for given date - targetHour: %d, date: %s, url: %s", hour, date, url)
	} else {
		temp = weatherData.Hourly.Temperature[hour]
	}

	temperature := fmt.Sprintf("%d", int(temp))

	if weatherCode != 100 {
		logger.LogMessage("GET weather successful, temp: %s, hour: %d, date: %s, url: %s", temperature, hour, date, url)
	}
	return emoji, temperature, nil
}
