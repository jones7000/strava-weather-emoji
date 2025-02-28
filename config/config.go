package config

import (
	"encoding/json"
	"os"
)

var (
	cfg      Config
	filename = "config.json"
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

func GetConfig() (Config, string) {
	return cfg, filename
}

func SetConfig(newConfig Config) error {
	cfg = newConfig
	err := SaveConfig()
	if err != nil {
		return err
	}
	return nil
}

func ReadConfig() error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return err
	}

	return nil
}

func SaveConfig() error {
	updatedJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, updatedJSON, 0644)
}
