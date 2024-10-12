package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	BlsEmail    string
	BlsPassword string

	ChatApiKey      string
	ProxyRow        string
	ProxyRowForeign string
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		return Config{}, err
	}

	email := os.Getenv("BLS_EMAIL")
	password := os.Getenv("BLS_PASSWORD")
	apiKey := os.Getenv("CHAT_API_KEY")
	proxyRow := os.Getenv("PROXY_ROW")
	proxyRowForeign := os.Getenv("PROXY_ROW_FOREIGN")

	config := Config{
		BlsEmail:        email,
		BlsPassword:     password,
		ChatApiKey:      apiKey,
		ProxyRow:        proxyRow,
		ProxyRowForeign: proxyRowForeign,
	}

	return config, nil
}
