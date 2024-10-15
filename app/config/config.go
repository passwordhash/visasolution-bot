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

	ImgurClientId     string
	ImgurClientSecret string
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		return Config{}, err
	}

	return Config{
		BlsEmail:          os.Getenv("BLS_EMAIL"),
		BlsPassword:       os.Getenv("BLS_PASSWORD"),
		ChatApiKey:        os.Getenv("CHAT_API_KEY"),
		ProxyRow:          os.Getenv("PROXY_ROW"),
		ProxyRowForeign:   os.Getenv("PROXY_ROW_FOREIGN"),
		ImgurClientId:     os.Getenv("IMGUR_CLIENT_ID"),
		ImgurClientSecret: os.Getenv("IMGUR_CLIENT_SECRET"),
	}, nil
}
