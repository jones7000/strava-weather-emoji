package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"strava-api/client/strava"
	"strava-api/client/weather"
	"strava-api/config"
	"strava-api/logger"
	"strava-api/model"
)

var cfg config.Config

func transformDateTime(activity model.ActivityResponse) (string, int, error) {

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

	logger.LogMessage("Using date: %s, calculated fullHour: %d", activity.StartDateLocal, fullHour)

	// Rückgabe des Datums und der vollen Stunde
	return date, fullHour, nil
}

func updateActivity(activityID string) {
	activity, err := strava.FetchActivityData(activityID)
	if err != nil {
		logger.LogMessage("Error fetchActivityData: %v", err)
		return
	}

	// check if indoor activitiy
	for _, indoor := range model.IndoorActivities {
		if indoor == activity.Type {
			logger.LogMessage("Indoor activity: %s type: %s, id: %s", activity.Name, activity.Type, activityID)
			return
		}
	}

	if len(activity.StartLatLon) == 0 {
		logger.LogMessage("No Lat Lon available in activity: %s %s", activity.Name, activityID)
		return
	}

	// transform activity data
	date, targetHour, err := transformDateTime(activity)
	if err != nil {
		logger.LogMessage("Error transformPolylineToLatLong: %v", err)
		return
	}

	// get weather emoji based on activity date, hour
	emoji, temp, err := weather.GetWeatherEmojiAndTemp(activity, date, targetHour)
	if err != nil {
		logger.LogMessage("Error getWeatherEmojiAndTemp %v", err)
		return
	}
	// send activity update to strava
	err = strava.SendActivityUpdate(activityID, activity, emoji, temp)
	if err != nil {
		logger.LogMessage("Error getWeatherEmoji %v", err)
		return
	}
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Strava sendet eine GET-Anfrage zur Verifizierung
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")

		if mode == "subscribe" && token == cfg.WebhookToken {
			logger.LogMessage("---> WEBHOOK VERIFIED <---")
			response := map[string]string{"hub.challenge": challenge}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			logger.LogMessage("---> Verifizierung fehlgeschlagen <---")
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
		return
	}

	if r.Method == http.MethodPost {
		var callback model.WebhookCallback
		err := json.NewDecoder(r.Body).Decode(&callback)
		if err != nil {
			http.Error(w, "Fehler beim Parsen des JSON", http.StatusBadRequest)
			return
		}

		logger.LogMessage("-------------------------------------------------------------")
		logger.LogMessage("received new webhook: %+v", callback)

		// Falls es sich um eine neue Aktivität handelt, rufe fetchActivityData auf
		if callback.AspectType == "create" && callback.ObjectType == "activity" {
			activityID := strconv.Itoa(callback.ObjectID)
			go updateActivity(activityID) // Asynchron ausführen, blockiert nicht den Webhook
		} else {
			logger.LogMessage("no action required")
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// Unbekannte Methoden ablehnen
	http.Error(w, "Error", http.StatusMethodNotAllowed)
}

func init() {
	err := config.ReadConfig()
	if err != nil {
		fmt.Println("Error reading JSON file: ", err)
		os.Exit(1)
	}

	cfg, _ = config.GetConfig()
	err = logger.InitLogger(cfg.LogTarget, cfg.LogFile)
	if err != nil {
		log.Fatalf("Logger-Initialisierung fehlgeschlagen: %v", err)
	}

}

func main() {
	defer logger.CloseLogger()
	http.HandleFunc("/webhook", WebhookHandler) // `/webhook` für GET (Verifikation) und POST (Events)
	logger.LogMessage("-------------------------------------------------------------")
	logger.LogMessage("server running on port %s...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
}
