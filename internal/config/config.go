package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"timesheet/internal/logging"
)

type Config struct {
	Name             string `json:"name"`
	CompanyName      string `json:"companyName"`
	FreeSpeech       string `json:"FreeSpeech"`
	StartAPIServer   bool   `json:"startApiServer"`
	SendDocumentType string `json:"sendDocumentType"`
	SendToOthers     bool   `json:"sendToOthers"`
	RecipientEmail   string `json:"recipientEmail,omitempty"`
	SenderEmail      string `json:"senderEmail,omitempty"`
	ReplyToEmail     string `json:"replyToEmail,omitempty"`
	ResendAPIKey     string `json:"resendApiKey,omitempty"`
}

func GetStartAPIServer() bool {
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return false
	}

	var configData struct {
		Name             string `json:"name"`
		CompanyName      string `json:"companyName"`
		FreeSpeech       string `json:"FreeSpeech"`
		SendDocumentType string `json:"sendDocumentType"`
		StartApiServer   bool   `json:"startApiServer"`
		SendToOthers     bool   `json:"sendToOthers"`
		RecipientEmail   string `json:"recipientEmail"`
		SenderEmail      string `json:"senderEmail"`
		ReplyToEmail     string `json:"replyToEmail"`
		ResendApiKey     string `json:"resendApiKey"`
	}

	if err := json.Unmarshal(configFile, &configData); err != nil {
		fmt.Println("Error parsing config JSON:", err)
		return false
	}

	return configData.StartApiServer
}

func checkConfig() bool {
	// Check if the config file exists
	_, err := os.Stat("config.json")
	if err != nil {
		fmt.Println("Uh oh:", err)
		return false
	}
	fmt.Println("Config file found!")
	return true
}

// GetEmailConfig reads the configuration file and returns email-related settings
func GetEmailConfig() (name string, companysendToOthers bool, recipientEmail, senderEmail, replyToEmail, resendAPIKey string, err error) {
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		return "", false, "", "", "", "", fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		return "", false, "", "", "", "", fmt.Errorf("error parsing config JSON: %w", err)
	}

	return config.Name, config.SendToOthers, config.RecipientEmail,
		config.SenderEmail, config.ReplyToEmail, config.ResendAPIKey, nil
}

func GetDocumentType() string {
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Printf("error reading config file: %v", err)
		return ""
	}
	var config struct {
		SendDocumentType string `json:"sendDocumentType"`
	}
	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Printf("error parsing config JSON: %v", err)
		return ""
	}
	return config.SendDocumentType
}

func GetUserConfig() (name string, companyName string, freeSpeech string, err error) {
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		return "", "", "", fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		return "", "", "", fmt.Errorf("error parsing config JSON: %w", err)
	}

	return config.Name, config.CompanyName, config.FreeSpeech, nil
}

func RequireConfig() {
	configPath := GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		config := Config{
			StartAPIServer:   true,
			SendDocumentType: "pdf",
		}
		SaveConfig(config)
		logging.Log("Created new config file at %s", configPath)
	} else {
		logging.Log("Config file is found!")
	}
}

func GetConfigPath() string {
	return filepath.Join(".", "config.json")
}

func SaveConfig(config Config) {
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling config:", err)
		os.Exit(1)
	}

	err = os.WriteFile("config.json", configData, 0644)
	if err != nil {
		fmt.Println("Error writing config file:", err)
		os.Exit(1)
	}
	time.Sleep(1 * time.Second)
}
