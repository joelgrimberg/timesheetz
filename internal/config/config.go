package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"timesheet/internal/logging"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

// Runtime development mode flag
var runtimeDevMode bool

// Config represents the application configuration
type Config struct {
	StartAPIServer   bool   `json:"startAPIServer"`
	SendDocumentType string `json:"sendDocumentType"`
	Name            string `json:"name"`
	CompanyName     string `json:"companyName"`
	FreeSpeech      string `json:"freeSpeech"`
	SendToOthers    bool   `json:"sendToOthers"`
	RecipientEmail  string `json:"recipientEmail"`
	SenderEmail     string `json:"senderEmail"`
	ReplyToEmail    string `json:"replyToEmail"`
	ResendAPIKey    string `json:"resendAPIKey"`
	DevelopmentMode bool   `json:"developmentMode"`
}

// SetRuntimeDevMode sets the runtime development mode
func SetRuntimeDevMode(devMode bool) {
	runtimeDevMode = devMode
	logging.Log("Runtime development mode set to: %v", devMode)
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
		config := Config{
			StartAPIServer:   true,
			SendDocumentType: "pdf",
			DevelopmentMode:  false,
		}

		// Should we run in accessible mode?
		accessible, _ := strconv.ParseBool(os.Getenv("ACCESSIBLE"))

		form := huh.NewForm(
			huh.NewGroup(huh.NewNote().
				Title("Timesheetz™ Setup").
				Description("Welcome to _Timesheetz™_.\nA Unicorny way to manage your timesheetz\n\nAight, Be a 🦄! \n\n").
				Next(true).
				NextLabel("Next"),
			),

			// Basic configuration
			huh.NewGroup(
				huh.NewInput().
					Value(&config.Name).
					Title("What is your name?").
					Placeholder("Uni Corn").
					Description("We'll use this to personalize your experience."),

				huh.NewInput().
					Value(&config.CompanyName).
					Title("What is the name of your company?").
					Placeholder("Uni Corn").
					Description("Don't worry, we all serve a master."),

				huh.NewInput().
					Value(&config.FreeSpeech).
					Title("What else do you want to share (will be put below the company name)").
					Placeholder("Uni Corn").
					Description("Free Speech"),

				huh.NewConfirm().
					Title("Do you want to start the API server every time you start the app?").
					Value(&config.StartAPIServer).
					Affirmative("Yes").
					Negative("No"),

				huh.NewConfirm().
					Title("Do you want to enable development mode?").
					Value(&config.DevelopmentMode).
					Affirmative("Yes").
					Negative("No").
					Description("Development mode uses a local database in the current directory."),
			),

			// Email configuration
			huh.NewGroup(
				huh.NewConfirm().
					Title("Would you like to be able to send the timesheet to someone who loves corny timesheetz?").
					Value(&config.SendToOthers).
					Affirmative("Yes").
					Negative("No"),
			),

			// Conditional email-related questions
			huh.NewGroup(
				huh.NewInput().
					Value(&config.RecipientEmail).
					Title("What is the recipient's email address?").
					Placeholder("recipient@example.com").
					Validate(func(s string) error {
						if s == "" && config.SendToOthers {
							return fmt.Errorf("email address is required")
						}
						return nil
					}),

				huh.NewInput().
					Value(&config.SenderEmail).
					Title("What is your email address?").
					Placeholder("you@example.com").
					Validate(func(s string) error {
						if s == "" && config.SendToOthers {
							return fmt.Errorf("email address is required")
						}
						return nil
					}),

				huh.NewInput().
					Value(&config.ReplyToEmail).
					Title("What is your reply-to email address?").
					Placeholder("you@example.com").
					Validate(func(s string) error {
						if s == "" && config.SendToOthers {
							return fmt.Errorf("email address is required")
						}
						return nil
					}),

				huh.NewInput().
					Value(&config.ResendAPIKey).
					Title("What is your Resend API key?").
					Placeholder("re_123456789").
					Password(true).
					Validate(func(s string) error {
						if s == "" && config.SendToOthers {
							return fmt.Errorf("Resend API key is required")
						}
						return nil
					}),
			).WithHideFunc(func() bool {
				return !config.SendToOthers
			}),
		).WithAccessible(accessible)

		err := form.Run()
		if err != nil {
			fmt.Println("Uh oh:", err)
			os.Exit(1)
		}

		// Prepare and write config
		prepareConfig := func() {
			SaveConfig(config)
		}

		_ = spinner.New().Title("Writing your configuration...").Accessible(accessible).Action(prepareConfig).Run()

		// Print config summary
		{
			style := lipgloss.NewStyle().
				Width(50).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(1, 2)

			titleStyle := lipgloss.NewStyle().Bold(true)
			highlight := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

			summary := fmt.Sprintf("%s\n\n", titleStyle.Render("CORNIFIGURATION COMPLETE"))
			summary += fmt.Sprintf("Name: %s\n", highlight.Render(config.Name))
			summary += fmt.Sprintf("Company: %s\n", highlight.Render(config.CompanyName))
			summary += fmt.Sprintf("Free Speech: %s\n", highlight.Render(config.FreeSpeech))
			summary += fmt.Sprintf("Start API Server: %s\n", highlight.Render(strconv.FormatBool(config.StartAPIServer)))
			summary += fmt.Sprintf("Development Mode: %s\n", highlight.Render(strconv.FormatBool(config.DevelopmentMode)))
			summary += fmt.Sprintf("Send to Others: %s\n", highlight.Render(strconv.FormatBool(config.SendToOthers)))

			if config.SendToOthers {
				summary += fmt.Sprintf("Recipient Email: %s\n", highlight.Render(config.RecipientEmail))
				summary += fmt.Sprintf("Sender Email: %s\n", highlight.Render(config.SenderEmail))
				summary += fmt.Sprintf("Reply-To Email: %s\n", highlight.Render(config.ReplyToEmail))
				summary += fmt.Sprintf("Resend API Key: %s\n", highlight.Render("****"+config.ResendAPIKey[len(config.ResendAPIKey)-4:]))
			}

			summary += "\nUnicornfiguration has been saved to config.json"
			fmt.Println(style.Render(summary))
		}

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

// GetDevelopmentMode returns whether development mode is enabled
func GetDevelopmentMode() bool {
	// Check runtime flag first
	if runtimeDevMode {
		return true
	}

	// Fall back to config file
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Printf("error reading config file: %v", err)
		return false
	}
	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Printf("error parsing config JSON: %v", err)
		return false
	}
	return config.DevelopmentMode
}
