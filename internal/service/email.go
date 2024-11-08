package service

import (
	"crypto/tls"
	"fmt"
	gomail "gopkg.in/mail.v2"
	"os"
	"visasolution/pkg/util"
)

const (
	emailTemplateFilename = "availbility-email-template.html"
	emailSubject          = "VisaSolution| Запись на подачу документов"
	assetsFolder          = "assets/"
	screenshotCID         = "12345"
)

type EmailDeps struct {
	Host     string
	Port     int
	Username string
	Password string

	ScreenshotFilePath string
}

type EmailService struct {
	d EmailDeps
}

func NewEmailService(d EmailDeps) *EmailService {
	return &EmailService{d: d}
}

func (e *EmailService) SendAvailbilityNotification(to string) error {
	screenshotFullPath := util.GetAbsolutePath(e.d.ScreenshotFilePath)

	emailTemplate, err := e.getEmailTemplate()
	if err != nil {
		return fmt.Errorf("error reading email template: %v", err)
	}
	emailTemplate = fmt.Sprintf(emailTemplate, screenshotCID)

	m := gomail.NewMessage()

	// прикрепляет картинку как вложение
	m.Attach(screenshotFullPath)
	m.SetHeader("From", "visasolution@passwordhash.tech")
	m.SetHeader("To", to)
	m.SetHeader("Subject", emailSubject)
	m.Embed(screenshotFullPath, gomail.SetHeader(map[string][]string{
		"Content-ID": {"<" + screenshotCID + ">"},
	}))
	m.SetBody("text/html", emailTemplate)

	d := gomail.NewDialer(e.d.Host, e.d.Port, e.d.Username, e.d.Password)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)
}

func (e *EmailService) getEmailTemplate() (string, error) {
	path := assetsFolder + emailTemplateFilename
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
