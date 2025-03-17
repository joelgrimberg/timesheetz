package db

import (
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	// Define test cases
	tests := []struct {
		name    string
		dbUser  string
		dbPass  string
		wantErr bool
	}{
		{"ValidCredentials", "root", "Shelter1!", false},
		{"InvalidCredentials", "invaliduser", "invalidpassword", true},
		{"EmptyUsernameAndPassword", "", "", true},
		{"ValidUsernameEmptyPassword", "root", "", true},
		{"EmptyUsernameValidPassword", "", "Shelter1!", true},
		{"SpecialCharacters", "user!@#", "pass!@#", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for the test
			os.Setenv("DBUSER", tt.dbUser)
			os.Setenv("DBPASSWORD", tt.dbPass)

			// Call the function to test
			err := Connect(tt.dbUser, tt.dbPass)

			// Check if the error matches the expected result
			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error: %v, but got: %v", tt.wantErr, err)
			}

			// Check if the db variable is not nil when no error is expected
			if !tt.wantErr && db == nil {
				t.Fatal("Expected db to be initialized, but it is nil")
			}
		})
	}
}
