package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	SeleniumUrl string

	BlsEmail    string
	BlsPassword string

	ChatApiKey   string
	ProxyForeign Proxy

	ImgurClientId     string
	ImgurClientSecret string

	SmtpHost     string
	SmtpPort     int
	SmtpUsername string
	Password     string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse smtp port: %w", err)
	}

	proxyRowForeign := os.Getenv("PROXY_ROW_FOREIGN")
	proxyForeign, err := ParseProxy(proxyRowForeign)
	if err != nil {
		return nil, fmt.Errorf("failed to parse foregin proxy row: %w", err)
	}

	return &Config{
		SeleniumUrl:       os.Getenv("SELENIUM_URL"),
		BlsEmail:          os.Getenv("BLS_EMAIL"),
		BlsPassword:       os.Getenv("BLS_PASSWORD"),
		ChatApiKey:        os.Getenv("CHAT_API_KEY"),
		ProxyForeign:      proxyForeign,
		ImgurClientId:     os.Getenv("IMGUR_CLIENT_ID"),
		ImgurClientSecret: os.Getenv("IMGUR_CLIENT_SECRET"),
		SmtpHost:          os.Getenv("SMTP_HOST"),
		SmtpPort:          smtpPort,
		SmtpUsername:      os.Getenv("SMTP_USERNAME"),
		Password:          os.Getenv("SMTP_PASSWORD"),
	}, nil
}
