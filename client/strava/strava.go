package strava

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"strava-api/config"
	"strava-api/logger"
	"strava-api/model"
)

func containsRunningActivity(activity string) bool {
	for _, a := range model.RunningActivities {
		if a == activity {
			return true
		}
	}
	return false
}

func SendActivityUpdate(activityID string, activity model.ActivityResponse, emoji string, temp string) error {
	cfg, cfgFile := config.GetConfig()
	token, err := RefreshToken(cfgFile, cfg)
	if err != nil {
		return fmt.Errorf("error beim Aktualisieren des Tokens: %v", err)
	}

	apiURL := cfg.APIUrlBase + "activities/" + activityID
	newName := activity.Name + " " + emoji
	logger.LogMessage("Update name: %s", newName)

	newDescription := activity.Description

	var running = containsRunningActivity(activity.Type)

	if temp != "999" && running {
		newDescription = activity.Description + fmt.Sprintf("T: %s°C", temp)
		logger.LogMessage("Update description: %s", newDescription)
	} else {
		logger.LogMessage("No description update, temp: %s, containsRunningActivity: %v", temp, running)
	}

	logger.LogMessage("Send activity id: %s", activityID)

	// Request-Daten erstellen
	updateRequest := model.ActivityResponse{
		Name:        newName,
		Description: newDescription,
	}

	// Request-Body in JSON umwandeln
	requestBody, err := json.Marshal(updateRequest)
	if err != nil {
		return fmt.Errorf("error beim Marshaling der Anfrage: %v", err)
	}

	// Neue PUT-Anfrage erstellen
	req, err := http.NewRequest("PUT", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error beim Erstellen der Anfrage: %v", err)
	}

	// Header setzen
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Anfrage ausführen
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error beim Ausführen des Requests: %v", err)
	}
	defer resp.Body.Close()

	// HTTP-Statuscode prüfen
	if resp.StatusCode != http.StatusOK {
		// Versuchen, die errormeldung aus dem Body zu lesen
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("error: HTTP-Status %d (Antwort konnte nicht gelesen werden)", resp.StatusCode)
		}
		return fmt.Errorf("error: HTTP-Status %d, Antwort: %s", resp.StatusCode, string(bodyBytes))
	}

	logger.LogMessage("Activity %s successfully updated: %s", activityID, newName)
	return nil
}

func FetchActivityData(activityID string) (model.ActivityResponse, error) {
	cfg, cfgFile := config.GetConfig()
	token, err := RefreshToken(cfgFile, cfg)
	if err != nil {
		return model.ActivityResponse{}, fmt.Errorf("error refreshing token")
	}

	apiURL := cfg.APIUrlBase + "activities/" + activityID

	// HTTP-GET-Anfrage ausführen
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return model.ActivityResponse{}, fmt.Errorf("error creating GET request")
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.ActivityResponse{}, fmt.Errorf("error GET Actitvity: %v", err)
	}
	defer resp.Body.Close()

	// Prüfen, ob der Statuscode OK ist (200)
	if resp.StatusCode != http.StatusOK {
		return model.ActivityResponse{}, fmt.Errorf("error unexpected HTTP code: %d", resp.StatusCode)
	}

	// Antwort auslesen
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.ActivityResponse{}, fmt.Errorf("error reading body")
	}

	// Das JSON in eine Struktur umwandeln
	var activity model.ActivityResponse
	err = json.Unmarshal(body, &activity)
	if err != nil {
		return model.ActivityResponse{}, fmt.Errorf("error parsing response %v", err)
	}

	logger.LogMessage("GET strava activity successful: %s", apiURL)

	return activity, nil
}

func RefreshToken(filename string, cfg config.Config) (string, error) {
	// Check if the token is expired
	currentTime := time.Now().Unix()
	if cfg.ExpiresAt > currentTime {
		expirationTime := time.Unix(cfg.ExpiresAt, 0).Format(time.RFC3339)
		logger.LogMessage("token is still valid. Expires at: %s", expirationTime)
		return cfg.AccessToken, nil
	}
	logger.LogMessage("refresh token.")

	url := cfg.APIUrlBase + "oauth/token"

	// Request body
	data := map[string]string{
		"client_id":     cfg.ClientID,
		"client_secret": cfg.ClientSecret,
		"grant_type":    "refresh_token",
		"refresh_token": cfg.RefreshToken,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// HTTP request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse response
	var tokenResponse model.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return "", err
	}

	// Update config with new tokens
	cfg.AccessToken = tokenResponse.AccessToken
	cfg.RefreshToken = tokenResponse.RefreshToken
	cfg.ExpiresAt = tokenResponse.ExpiresAt

	// Save updated config
	err = config.SetConfig(cfg)
	if err != nil {
		return "", err
	}
	logger.LogMessage("token successfully refreshed.")
	return cfg.AccessToken, nil
}
