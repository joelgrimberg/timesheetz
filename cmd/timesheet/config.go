package main

import (
	"log"
	"os"
)

func GetDBCredentials() (string, string) {
	dbUser := os.Getenv("DBUSER")
	if dbUser == "" {
		log.Fatal("DBUSER environment variable is not set")
	}

	dbPassword := os.Getenv("DBPASSWORD")
	if dbPassword == "" {
		log.Fatal("DBPASSWORD environment variable is not set")
	}

	return dbUser, dbPassword
}
