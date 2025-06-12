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
)

// Runtime development mode flag
var runtimeDevMode bool
var runtimePort int

// TrainingHours represents the training hours configuration
type TrainingHours struct {
	YearlyTarget int    `json:"yearlyTarget"`
	Category     string `json:"category"`
}

// Config represents the application configuration
type Config struct {
	// User Information
	Name        string `json:"name"`
	CompanyName string `json:"companyName"`
	FreeSpeech  string `json:"freeSpeech"`

	// API Server Configuration
	StartAPIServer bool `json:"startAPIServer"`
	APIPort        int  `json:"apiPort"`

	// Development Settings
	DevelopmentMode bool `json:"developmentMode"`

	// Document Settings
	SendDocumentType string `json:"sendDocumentType"`

	// Email Configuration
	SendToOthers   bool   `json:"sendToOthers"`
	RecipientEmail string `json:"recipientEmail"`
	SenderEmail    string `json:"senderEmail"`
	ReplyToEmail   string `json:"replyToEmail"`
	ResendAPIKey   string `json:"resendApiKey"`

	// Training Hours Configuration
	TrainingHours TrainingHours `json:"trainingHours"`
}

// SetRuntimeDevMode sets the runtime development mode
func SetRuntimeDevMode(devMode bool) {
	runtimeDevMode = devMode
	logging.Log("Runtime development mode set to: %v", devMode)
}

// SetRuntimePort sets the runtime API port
func SetRuntimePort(port int) {
	runtimePort = port
	logging.Log("Runtime API port set to: %v", port)
}

// GetAPIPort returns the API port to use
func GetAPIPort() int {
	// Check runtime flag first
	if runtimePort != 0 {
		return runtimePort
	}

	// Fall back to config file
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("Error: No port specified. Please either:")
		fmt.Println("  1. Add 'apiPort' to your config.json file")
		fmt.Println("  2. Run the program with --port flag")
		fmt.Println("  3. Run the program with --no-tui flag if you don't need the API server")
		os.Exit(1)
	}
	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		fmt.Println("Error: Invalid config.json file. Please check your configuration.")
		os.Exit(1)
	}
	if config.APIPort == 0 {
		fmt.Println("Error: No port specified. Please either:")
		fmt.Println("  1. Add 'apiPort' to your config.json file")
		fmt.Println("  2. Run the program with --port flag")
		fmt.Println("  3. Run the program with --no-tui flag if you don't need the API server")
		os.Exit(1)
	}
	return config.APIPort
}

func GetStartAPIServer() bool {
	configPath := GetConfigPath()
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return false
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		fmt.Println("Error parsing config JSON:", err)
		return false
	}

	return config.StartAPIServer
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
	configPath := GetConfigPath()
	configFile, err := os.ReadFile(configPath)
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
	configPath := GetConfigPath()
	configFile, err := os.ReadFile(configPath)
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
	configPath := GetConfigPath()
	configFile, err := os.ReadFile(configPath)
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
			// User Information
			Name:        "",
			CompanyName: "",
			FreeSpeech:  "",

			// API Server Configuration
			StartAPIServer: true,
			APIPort:        8080,

			// Development Settings
			DevelopmentMode: false,

			// Document Settings
			SendDocumentType: "pdf",

			// Email Configuration
			SendToOthers:   false,
			RecipientEmail: "",
			SenderEmail:    "",
			ReplyToEmail:   "",
			ResendAPIKey:   "",

			// Training Hours Configuration
			TrainingHours: TrainingHours{
				YearlyTarget: 36, // Default to 36 hours
				Category:     "Training",
			},
		}

		// Should we run in accessible mode?
		accessible, _ := strconv.ParseBool(os.Getenv("ACCESSIBLE"))

		// Create a string variable for port input
		portStr := "8080"
		trainingHoursStr := "36"

		form := huh.NewForm(
			huh.NewGroup(huh.NewNote().
				Title("Timesheetz™ Setup").
				Description("Welcome to _Timesheetz™_.\nA Unicorny way to manage your timesheetz\n\nAight, Be a 🦄! \n\n").
				Next(true).
				NextLabel("Next"),
			),

			// User Information
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
			),

			// Training Hours Configuration
			huh.NewGroup(
				huh.NewInput().
					Value(&trainingHoursStr).
					Title("How many training hours are allocated per year?").
					Placeholder("36").
					Description("This is the total number of training hours you can use per year."),
			),

			// API Server Configuration
			huh.NewGroup(
				huh.NewConfirm().
					Title("Do you want to start the API server every time you start the app?").
					Value(&config.StartAPIServer).
					Affirmative("Yes").
					Negative("No"),

				huh.NewInput().
					Value(&portStr).
					Title("What port should the API server run on?").
					Placeholder("8080").
					Validate(func(s string) error {
						port, err := strconv.Atoi(s)
						if err != nil {
							return fmt.Errorf("port must be a number")
						}
						if port < 1 || port > 65535 {
							return fmt.Errorf("port must be between 1 and 65535")
						}
						return nil
					}),
			),

			// Development Settings
			huh.NewGroup(
				huh.NewConfirm().
					Title("Do you want to enable development mode?").
					Value(&config.DevelopmentMode).
					Affirmative("Yes").
					Negative("No").
					Description("Development mode uses a local database in the current directory."),
			),

			// Document Settings
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What document type do you want to use for exports?").
					Options(
						huh.NewOption("PDF", "pdf"),
						huh.NewOption("Excel", "excel"),
					).
					Value(&config.SendDocumentType),
			),

			// Email Configuration
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

			// Save the configuration
			huh.NewGroup(
				huh.NewNote().
					Title("Saving Configuration").
					Description("Saving your configuration..."),
			),
		).WithAccessible(accessible)

		err := form.Run()
		if err != nil {
			fmt.Println("Error running form:", err)
			os.Exit(1)
		}

		// Convert port string to integer
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Println("Error: Invalid port number")
			os.Exit(1)
		}
		config.APIPort = port

		// Convert training hours string to integer
		trainingHours, err := strconv.Atoi(trainingHoursStr)
		if err != nil {
			fmt.Println("Error: Invalid training hours number")
			os.Exit(1)
		}
		config.TrainingHours.YearlyTarget = trainingHours

		// Save the configuration
		SaveConfig(config)
	} else {
		logging.Log("Config file is found!")
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory if UserConfigDir fails
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get user home directory: %v", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "timesheetz", "config.json")
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config Config) {
	configPath := GetConfigPath()

	// Ensure the directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Println("Error creating config directory:", err)
		os.Exit(1)
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling config:", err)
		os.Exit(1)
	}

	err = os.WriteFile(configPath, configData, 0644)
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
		logging.Log("Development mode enabled via runtime flag")
		return true
	}

	// Fall back to config file
	configPath := GetConfigPath()
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("error reading config file: %v", err)
		return false
	}
	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Printf("error parsing config JSON: %v", err)
		return false
	}

	if config.DevelopmentMode {
		logging.Log("Development mode enabled via config file")
		return true
	}

	logging.Log("Development mode disabled")
	return false
}

// GetConfig reads and returns the configuration from the config file
func GetConfig() (Config, error) {
	configPath := GetConfigPath()
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}
