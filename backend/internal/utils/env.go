package utils

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvIfNeeded() {
	if err := godotenv.Load(".env"); err != nil {
		if err2 := godotenv.Load("../../.env"); err2 != nil {
			log.Fatalf("No .env file found at .env or ../../.env, and loading env failed: %v, %v", err, err2)
		}
	}
}
