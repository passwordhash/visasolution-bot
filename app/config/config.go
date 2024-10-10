package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	BlsEmail    string
	BlsPassword string

	ChatApiKey string
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		return Config{}, err
	}

	email := os.Getenv("BLS_EMAIL")
	password := os.Getenv("BLS_PASSWORD")
	apiKey := os.Getenv("CHAT_API_KEY")

	config := Config{
		BlsEmail:    email,
		BlsPassword: password,
		ChatApiKey:  apiKey,
	}

	return config, nil
}
