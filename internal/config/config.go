package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	SeleniumUrl string

	BlsEmail    string
	BlsPassword string

	ChatApiKey      string
	ProxyRow        string
	ProxyRowForeign string

	ImgurClientId     string
	ImgurClientSecret string

	SmtpHost     string
	SmtpPort     int
	SmtpUsername string
	Password     string
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		return Config{}, err
	}

	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))

	return Config{
		SeleniumUrl:       os.Getenv("SELENIUM_URL"),
		BlsEmail:          os.Getenv("BLS_EMAIL"),
		BlsPassword:       os.Getenv("BLS_PASSWORD"),
		ChatApiKey:        os.Getenv("CHAT_API_KEY"),
		ProxyRow:          os.Getenv("PROXY_ROW_RUSSIA"),
		ProxyRowForeign:   os.Getenv("PROXY_ROW_FOREIGN"),
		ImgurClientId:     os.Getenv("IMGUR_CLIENT_ID"),
		ImgurClientSecret: os.Getenv("IMGUR_CLIENT_SECRET"),
		SmtpHost:          os.Getenv("SMTP_HOST"),
		SmtpPort:          smtpPort,
		SmtpUsername:      os.Getenv("SMTP_USERNAME"),
		Password:          os.Getenv("SMTP_PASSWORD"),
	}, nil
}
