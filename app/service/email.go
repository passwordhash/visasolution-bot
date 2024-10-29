package service

import (
	"crypto/tls"
	"fmt"
	gomail "gopkg.in/mail.v2"
	"os"
	"visasolution/app/util"
)

const (
	emailTemplateFilename = "availbility-email-template.html"
	emailSubject          = "VisaSolution| Запись на подачу документов"
	assetsFolder          = "assets"
)

type EmailDeps struct {
	Host     string
	Port     int
	Username string
	Password string
}

type EmailService struct {
	d EmailDeps
}

func NewEmailService(d EmailDeps) *EmailService {
	return &EmailService{d: d}
}

func (e *EmailService) SendAvailbilityNotification(to string) error {
	availabilityEmailTemplate, err := e.getEmailTemplate()
	if err != nil {
		return fmt.Errorf("error reading email template: %v", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "visasolution@passwordhash.tech")
	m.SetHeader("To", to)
	m.SetHeader("Subject", emailSubject)
	m.SetBody("text/html", availabilityEmailTemplate)

	d := gomail.NewDialer(e.d.Host, e.d.Port, e.d.Username, e.d.Password)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)
}

func (e *EmailService) getEmailTemplate() (string, error) {
	path := util.GetAbsolutePath(assetsFolder + "/" + emailTemplateFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
