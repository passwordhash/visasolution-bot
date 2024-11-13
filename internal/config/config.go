package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	NotifiedEmails    []string
	MainLoopIntervalM int

	BlsEmail    string
	BlsPassword string
	SeleniumUrl string

	ChatApiKey string

	ImgurClientId     string
	ImgurClientSecret string

	SmtpHost     string
	SmtpPort     int
	SmtpUsername string
	Password     string
}

const defaultMainLoopIntervalM = 30

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	notifiedEmails := os.Getenv("NOTIFIED_EMAILS")
	if notifiedEmails == "" {
		return nil, fmt.Errorf(".env NOTIFIED_EMAILS is required")
	}

	mainLoopIntervalM, err := strconv.Atoi(os.Getenv("MAIN_LOOP_INTERVAL_M"))
	if err != nil || mainLoopIntervalM <= 0 {
		mainLoopIntervalM = defaultMainLoopIntervalM
	}

	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse smtp port: %w", err)
	}

	return &Config{
		NotifiedEmails:    strings.Split(notifiedEmails, ";"),
		MainLoopIntervalM: mainLoopIntervalM,
		SeleniumUrl:       os.Getenv("SELENIUM_URL"),
		BlsEmail:          os.Getenv("BLS_EMAIL"),
		BlsPassword:       os.Getenv("BLS_PASSWORD"),
		ChatApiKey:        os.Getenv("CHAT_API_KEY"),
		ImgurClientId:     os.Getenv("IMGUR_CLIENT_ID"),
		ImgurClientSecret: os.Getenv("IMGUR_CLIENT_SECRET"),
		SmtpHost:          os.Getenv("SMTP_HOST"),
		SmtpPort:          smtpPort,
		SmtpUsername:      os.Getenv("SMTP_USERNAME"),
		Password:          os.Getenv("SMTP_PASSWORD"),
	}, nil
}
