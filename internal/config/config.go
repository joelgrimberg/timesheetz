package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"timesheet/internal/logging"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

// Runtime development mode flag
var runtimeDevMode bool
var runtimePort int
var runtimeDBType string
var runtimePostgresURL string

// TrainingHours represents the training hours configuration
type TrainingHours struct {
	YearlyTarget int    `json:"yearlyTarget"`
	Category     string `json:"category"`
}

// VacationHours represents the vacation hours configuration
type VacationHours struct {
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

	// API Client Configuration (for remote mode)
	APIMode    string `json:"apiMode"`    // "local", "dual", or "remote" (default: "local")
	APIBaseURL string `json:"apiBaseURL"` // Base URL for remote API (e.g., "http://timesheetz.local")

	// Database Configuration
	DBLocation  string `json:"dbLocation"`
	DBType      string `json:"dbType"`      // "sqlite" (default) or "postgres"
	PostgresURL string `json:"postgresURL"` // PostgreSQL connection string

	// Development Settings
	DevelopmentMode bool `json:"developmentMode"`

	// Document Settings
	SendDocumentType string `json:"sendDocumentType"`
	ExportLanguage   string `json:"exportLanguage"` // "en" or "nl" (default: "en")

	// Email Configuration
	SendToOthers   bool   `json:"sendToOthers"`
	RecipientEmail string `json:"recipientEmail"`
	SenderEmail    string `json:"senderEmail"`
	ReplyToEmail   string `json:"replyToEmail"`
	ResendAPIKey   string `json:"resendApiKey"`

	// Training Hours Configuration
	TrainingHours TrainingHours `json:"trainingHours"`

	// Vacation Hours Configuration
	VacationHours VacationHours `json:"vacationHours"`
}

// SetRuntimeDevMode sets the runtime development mode
func SetRuntimeDevMode(devMode bool) {
	runtimeDevMode = devMode
	logging.Log("Runtime development mode set to: %v", devMode)
}

// SetRuntimePort sets the runtime API port
func SetRuntimePort(port int) {
	runtimePort = port
	// Use fmt.Printf directly to avoid potential logging issues
	if logging.IsVerbose() {
		fmt.Printf("Runtime API port set to: %v\n", port)
	}
	logging.Log("Runtime API port set to: %v", port)
}

// GetAPIPort returns the API port to use
func GetAPIPort() int {
	// Check runtime flag first
	if runtimePort != 0 {
		return runtimePort
	}

	// Fall back to config file
	configPath := GetConfigPath()
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		// In non-interactive mode (like Docker), default to 8080 instead of exiting
		if os.Getenv("TIMESHEETZ_NO_TUI") == "true" || !term.IsTerminal(int(os.Stdin.Fd())) {
			logging.Log("Warning: Could not read config file, defaulting to port 8080")
			return 8080
		}
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

func GetExportLanguage() string {
	configPath := GetConfigPath()
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return "en"
	}
	var config struct {
		ExportLanguage string `json:"exportLanguage"`
	}
	if err := json.Unmarshal(configFile, &config); err != nil {
		return "en"
	}
	if config.ExportLanguage == "" {
		return "en"
	}
	return config.ExportLanguage
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
	logging.Log("Checking for config file at: %s", configPath)
	_, err := os.Stat(configPath)
	if err != nil {
		// Only show setup if file doesn't exist
		if os.IsNotExist(err) {
			// Check if we're in a non-interactive environment (like Docker)
			// Check multiple conditions: terminal, environment variable, or --no-tui flag
			isNonInteractive := !isTerminal(os.Stdin) || os.Getenv("TIMESHEETZ_NO_TUI") == "true"
			if isNonInteractive {
				logging.Log("Config file not found, but running in non-interactive mode. Creating default config...")
				config := Config{
					// User Information
					Name:        "",
					CompanyName: "",
					FreeSpeech:  "",

					// API Server Configuration
					StartAPIServer: true,
					APIPort:        8080,

					// API Client Configuration
					APIMode:    "local", // Default to local mode
					APIBaseURL: "",      // Empty means use local database

					// Database Location
					DBLocation: "",

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

					// Vacation Hours Configuration
					VacationHours: VacationHours{
						YearlyTarget: 0, // Default to 0 hours
						Category:     "Vacation",
					},
				}
				SaveConfig(config)
				logging.Log("Default config created successfully")
				return
			}
			logging.Log("Config file not found, showing setup form...")
			config := Config{
				// User Information
				Name:        "",
				CompanyName: "",
				FreeSpeech:  "",

				// API Server Configuration
				StartAPIServer: true,
				APIPort:        8080,

				// API Client Configuration
				APIMode:    "local", // Default to local mode
				APIBaseURL: "",      // Empty means use local database

				// Database Location
				DBLocation: "",

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

				// Vacation Hours Configuration
				VacationHours: VacationHours{
					YearlyTarget: 0, // Default to 0 hours
					Category:     "Vacation",
				},
			}

			// Should we run in accessible mode?
			accessible, _ := strconv.ParseBool(os.Getenv("ACCESSIBLE"))

			// Create a string variable for port input
			portStr := "8080"
			trainingHoursStr := "36"
			vacationHoursStr := "0"
			dbLocationStr := ""

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

				// Database Configuration
				huh.NewGroup(
					huh.NewInput().
						Value(&dbLocationStr).
						Title("Where should your database be stored?").
						Placeholder("/path/to/timesheet.db").
						Description("Leave empty to use the default location (~/.config/timesheetz/timesheet.db). You can specify a full path to store it elsewhere."),
				),

				// Training Hours Configuration
				huh.NewGroup(
					huh.NewInput().
						Value(&trainingHoursStr).
						Title("How many training hours are allocated per year?").
						Placeholder("36").
						Description("This is the total number of training hours you can use per year."),
				),

				// Vacation Hours Configuration
				huh.NewGroup(
					huh.NewInput().
						Value(&vacationHoursStr).
						Title("How many vacation hours are allocated per year?").
						Placeholder("0").
						Description("This is the total number of vacation hours you can use per year."),
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

			// Convert vacation hours string to integer
			vacationHours, err := strconv.Atoi(vacationHoursStr)
			if err != nil {
				fmt.Println("Error: Invalid vacation hours number")
				os.Exit(1)
			}
			config.VacationHours.YearlyTarget = vacationHours

			// Set database location (empty string means use default)
			config.DBLocation = dbLocationStr

			// Save the configuration
			SaveConfig(config)
		} else {
			// File exists but there's another error (permissions, etc.)
			logging.Log("Warning: Error checking config file at %s: %v", configPath, err)
			logging.Log("Continuing anyway...")
		}
	} else {
		// Config file exists and is accessible
		logging.Log("Config file found at: %s", configPath)
	}
}

// GetConfigPath returns the path to the config file
// Uses XDG Base Directory Specification: ~/.config/timesheetz/config.json
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}
	return filepath.Join(homeDir, ".config", "timesheetz", "config.json")
}

// SaveConfig saves the configuration to a file
func SaveConfig(config Config) error {
	configPath := GetConfigPath()

	// Ensure the directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
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

	// Create debug info
	debugInfo := map[string]interface{}{
		"configPath": configPath,
	}

	configFile, err := os.ReadFile(configPath)
	if err != nil {
		debugInfo["error"] = fmt.Sprintf("Error reading config file: %v", err)
		writeDebugToFile(debugInfo)
		return Config{}, err
	}

	debugInfo["configContent"] = string(configFile)

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		debugInfo["error"] = fmt.Sprintf("Error parsing config JSON: %v", err)
		writeDebugToFile(debugInfo)
		return Config{}, err
	}

	debugInfo["parsedVacationHours"] = config.VacationHours
	writeDebugToFile(debugInfo)

	return config, nil
}

// writeDebugToFile writes debug information to a JSON file
func writeDebugToFile(debugInfo map[string]interface{}) {
	debugJSON, err := json.MarshalIndent(debugInfo, "", "  ")
	if err != nil {
		return
	}

	// Write to debug file in the same directory as config
	configDir := filepath.Dir(GetConfigPath())
	debugPath := filepath.Join(configDir, "config_debug.json")
	os.WriteFile(debugPath, debugJSON, 0644)
}

// GetDBPath returns the path to the database file, using config if set
func GetDBPath() string {
	// Check environment variable first (useful for Docker/containerized deployments)
	if dbPath := os.Getenv("TIMESHEETZ_DB_PATH"); dbPath != "" {
		// Expand ~ in path if present
		if strings.HasPrefix(dbPath, "~/") {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				dbPath = filepath.Join(homeDir, dbPath[2:])
			}
		}
		return dbPath
	}

	// Check config file
	config, err := GetConfig()
	if err == nil && config.DBLocation != "" {
		// Expand ~ in path if present
		if strings.HasPrefix(config.DBLocation, "~/") {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				return filepath.Join(homeDir, config.DBLocation[2:])
			}
		}
		return config.DBLocation
	}

	// Default location
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory if UserConfigDir fails
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get user home directory: %v", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "timesheetz", "timesheet.db")
}

// GetAPIMode returns the API mode: "local", "dual", or "remote"
func GetAPIMode() string {
	// Check environment variable first
	if envMode := os.Getenv("TIMESHEETZ_API_MODE"); envMode != "" {
		if envMode == "local" || envMode == "dual" || envMode == "remote" {
			return envMode
		}
	}

	// Fall back to config file
	config, err := GetConfig()
	if err != nil {
		return "local" // Default to local mode
	}

	if config.APIMode == "" {
		return "local" // Default to local mode
	}

	if config.APIMode != "local" && config.APIMode != "dual" && config.APIMode != "remote" {
		logging.Log("Invalid apiMode '%s', defaulting to 'local'", config.APIMode)
		return "local"
	}

	return config.APIMode
}

// isTerminal checks if the given file descriptor is a terminal
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// GetAPIBaseURL returns the base URL for the remote API
func GetAPIBaseURL() string {
	// Check environment variable first
	if envURL := os.Getenv("TIMESHEETZ_API_URL"); envURL != "" {
		return envURL
	}

	// Fall back to config file
	config, err := GetConfig()
	if err != nil {
		return ""
	}

	// If apiMode is "local", return empty string (no remote API)
	if config.APIMode == "local" || config.APIMode == "" {
		return ""
	}

	// If apiBaseURL is set, use it
	if config.APIBaseURL != "" {
		return config.APIBaseURL
	}

	// If apiMode is "dual" or "remote" but no base URL, try to construct from port
	// This is a fallback for backward compatibility
	if config.APIPort != 0 {
		return fmt.Sprintf("http://localhost:%d", config.APIPort)
	}

	return ""
}

// SetRuntimeDBType sets the runtime database type
func SetRuntimeDBType(dbType string) {
	runtimeDBType = dbType
	logging.Log("Runtime database type set to: %v", dbType)
}

// SetRuntimePostgresURL sets the runtime PostgreSQL URL
func SetRuntimePostgresURL(url string) {
	runtimePostgresURL = url
	logging.Log("Runtime PostgreSQL URL set")
}

// GetDBType returns the database type: "sqlite" or "postgres"
func GetDBType() string {
	// Check runtime flag first (CLI)
	if runtimeDBType != "" {
		return runtimeDBType
	}

	// Check environment variable
	if envType := os.Getenv("TIMESHEETZ_DB_TYPE"); envType != "" {
		if envType == "sqlite" || envType == "postgres" {
			return envType
		}
		logging.Log("Invalid TIMESHEETZ_DB_TYPE '%s', defaulting to 'sqlite'", envType)
	}

	// Fall back to config file
	config, err := GetConfig()
	if err != nil {
		return "sqlite" // Default
	}

	if config.DBType == "" {
		return "sqlite"
	}

	if config.DBType != "sqlite" && config.DBType != "postgres" {
		logging.Log("Invalid dbType '%s' in config, defaulting to 'sqlite'", config.DBType)
		return "sqlite"
	}

	return config.DBType
}

// GetPostgresURL returns the PostgreSQL connection URL
func GetPostgresURL() string {
	// Check runtime flag first (CLI)
	if runtimePostgresURL != "" {
		return runtimePostgresURL
	}

	// Check environment variable
	if envURL := os.Getenv("TIMESHEETZ_POSTGRES_URL"); envURL != "" {
		return envURL
	}

	// Fall back to config file
	config, err := GetConfig()
	if err != nil {
		return ""
	}

	return config.PostgresURL
}
