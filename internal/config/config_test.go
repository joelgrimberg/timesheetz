package config

import (
	"log"
	"os"
	"testing"
)

// disableLogging temporarily disables logging during tests
func disableLogging() func() {
	// Save the current log output
	originalOutput := log.Writer()
	// Set log output to discard
	log.SetOutput(os.NewFile(0, os.DevNull))
	// Return a function to restore the original output
	return func() {
		log.SetOutput(originalOutput)
	}
}

func TestSaveAndGetUserConfig(t *testing.T) {
	// Disable logging for this test
	restoreLogging := disableLogging()
	defer restoreLogging()

	// Remove any existing config file to start fresh
	configPath := GetConfigPath()
	os.Remove(configPath)

	// Create a temporary config for testing
	testConfig := Config{
		Name:        "Test User",
		CompanyName: "Test Company",
		FreeSpeech:  "Test Speech",
	}

	// Save the test config
	SaveConfig(testConfig)

	// Clean up the test file after the test
	defer os.Remove(configPath)

	// Get the user config
	name, companyName, freeSpeech, err := GetUserConfig()
	if err != nil {
		t.Errorf("GetUserConfig failed: %v", err)
	}

	// Verify the values
	if name != testConfig.Name {
		t.Errorf("Expected name %q, got %q", testConfig.Name, name)
	}
	if companyName != testConfig.CompanyName {
		t.Errorf("Expected company name %q, got %q", testConfig.CompanyName, companyName)
	}
	if freeSpeech != testConfig.FreeSpeech {
		t.Errorf("Expected free speech %q, got %q", testConfig.FreeSpeech, freeSpeech)
	}
}

func TestGetAPIPort(t *testing.T) {
	// Disable logging for this test
	restoreLogging := disableLogging()
	defer restoreLogging()

	// Remove any existing config file to start fresh
	configPath := GetConfigPath()
	os.Remove(configPath)

	// Create a minimal config file with default port
	testConfig := Config{
		APIPort: 8080,
	}
	SaveConfig(testConfig)
	defer os.Remove(configPath)

	// Test default port from config
	port := GetAPIPort()
	if port != 8080 {
		t.Errorf("Expected default port 8080, got %d", port)
	}

	// Test custom port from config
	testConfig.APIPort = 3000
	SaveConfig(testConfig)

	port = GetAPIPort()
	if port != 3000 {
		t.Errorf("Expected port 3000, got %d", port)
	}

	// Test runtime port override
	SetRuntimePort(4000)
	port = GetAPIPort()
	if port != 4000 {
		t.Errorf("Expected runtime port 4000, got %d", port)
	}
	// Reset runtime port for other tests
	SetRuntimePort(0)
}

func TestGetStartAPIServer(t *testing.T) {
	// Disable logging for this test
	restoreLogging := disableLogging()
	defer restoreLogging()

	// Remove any existing config file to start fresh
	configPath := GetConfigPath()
	os.Remove(configPath)

	// Test default value when no config exists
	startServer := GetStartAPIServer()
	if startServer {
		t.Error("Expected default StartAPIServer to be false")
	}

	// Test custom value from config
	testConfig := Config{
		StartAPIServer: true,
	}
	SaveConfig(testConfig)
	defer os.Remove(configPath)

	startServer = GetStartAPIServer()
	if !startServer {
		t.Error("Expected StartAPIServer to be true")
	}
}

func TestGetDocumentType(t *testing.T) {
	// Disable logging for this test
	restoreLogging := disableLogging()
	defer restoreLogging()

	// Remove any existing config file to start fresh
	configPath := GetConfigPath()
	os.Remove(configPath)

	// Test default value when no config exists
	docType := GetDocumentType()
	if docType != "" {
		t.Errorf("Expected empty document type, got %q", docType)
	}

	// Test custom value from config
	testConfig := Config{
		SendDocumentType: "excel",
	}
	SaveConfig(testConfig)
	defer os.Remove(configPath)

	docType = GetDocumentType()
	if docType != "excel" {
		t.Errorf("Expected document type 'excel', got %q", docType)
	}
}

func TestGetEmailConfig(t *testing.T) {
	// Disable logging for this test
	restoreLogging := disableLogging()
	defer restoreLogging()

	// Remove any existing config file to start fresh
	configPath := GetConfigPath()
	os.Remove(configPath)

	// Test default values when no config exists
	name, sendToOthers, recipient, sender, replyTo, apiKey, err := GetEmailConfig()
	if err == nil {
		t.Error("Expected error when config file doesn't exist")
	}
	if name != "" || sendToOthers || recipient != "" || sender != "" || replyTo != "" || apiKey != "" {
		t.Error("Expected empty email config values")
	}

	// Test custom values from config
	testConfig := Config{
		Name:           "Test User",
		SendToOthers:   true,
		RecipientEmail: "recipient@test.com",
		SenderEmail:    "sender@test.com",
		ReplyToEmail:   "reply@test.com",
		ResendAPIKey:   "test_api_key",
	}
	SaveConfig(testConfig)
	defer os.Remove(configPath)

	name, sendToOthers, recipient, sender, replyTo, apiKey, err = GetEmailConfig()
	if err != nil {
		t.Errorf("GetEmailConfig failed: %v", err)
	}
	if name != "Test User" {
		t.Errorf("Expected name 'Test User', got %q", name)
	}
	if !sendToOthers {
		t.Error("Expected SendToOthers to be true")
	}
	if recipient != "recipient@test.com" {
		t.Errorf("Expected recipient 'recipient@test.com', got %q", recipient)
	}
	if sender != "sender@test.com" {
		t.Errorf("Expected sender 'sender@test.com', got %q", sender)
	}
	if replyTo != "reply@test.com" {
		t.Errorf("Expected reply-to 'reply@test.com', got %q", replyTo)
	}
	if apiKey != "test_api_key" {
		t.Errorf("Expected API key 'test_api_key', got %q", apiKey)
	}
}

func TestGetDevelopmentMode(t *testing.T) {
	// Disable logging for this test
	restoreLogging := disableLogging()
	defer restoreLogging()

	// Remove any existing config file to start fresh
	configPath := GetConfigPath()
	os.Remove(configPath)

	// Test default value when no config exists
	devMode := GetDevelopmentMode()
	if devMode {
		t.Error("Expected default development mode to be false")
	}

	// Test custom value from config
	testConfig := Config{
		DevelopmentMode: true,
	}
	SaveConfig(testConfig)
	defer os.Remove(configPath)

	devMode = GetDevelopmentMode()
	if !devMode {
		t.Error("Expected development mode to be true")
	}

	// Test runtime override
	SetRuntimeDevMode(true)
	devMode = GetDevelopmentMode()
	if !devMode {
		t.Error("Expected runtime development mode to be true")
	}
} 