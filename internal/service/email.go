package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	gomail "gopkg.in/mail.v2"
	"os"
	"visasolution/pkg/util"
)

const (
	emailFrom             = "visasolution@passwordhash.tech"
	emailTemplateFilename = "availbility-email-template.html"
	emailSubject          = "VisaSolution| Запись на подачу документов"
	assetsFolder          = "assets/"
	screenshotCID         = "12345"
)

type SendEmailsError struct {
	emailStats map[string]error
}

func (e SendEmailsError) Error() string {
	errMsg := bytes.NewBufferString("error sending emails: ")
	for email, err := range e.emailStats {
		errMsg.WriteString(fmt.Sprintf("\temail: %s, error: %v", email, err))
	}
	return errMsg.String()
}

type EmailDeps struct {
	Host     string
	Port     int
	Username string
	Password string

	ScreenshotFilePath string
}

type Emailer interface {
	SetTo(to string)
	Message() *gomail.Message
}

type EmailMessage struct {
	m *gomail.Message
}

func (e *EmailMessage) SetTo(to string) {
	e.m.SetHeader("To", to)
}

func (e *EmailMessage) Message() *gomail.Message {
	return e.m
}

type EmailService struct {
	d EmailDeps
}

func NewEmailService(d EmailDeps) *EmailService {
	return &EmailService{d: d}
}

// SendBulkNotification отправляет уведомление на указанные email.
// Возвращает список успешно отправленных email и ошибку.
// Если не удалось отправить email, то возвращается список успешно отправленных email и SendEmailsError.
func (e *EmailService) SendBulkNotification(to []string) ([]string, error) {
	sent := make([]string, 0, len(to))

	baseMsg, err := e.availabilityEmailMessage()
	if err != nil {
		return nil, fmt.Errorf("error creating email message: %v", err)
	}

	var sendErr SendEmailsError
	for _, email := range to {
		err := e.sendTo(baseMsg, email)
		if err == nil {
			sent = append(sent, email)
			continue
		}
		sendErr.emailStats[email] = err
	}

	if len(sent) != len(to) {
		return sent, sendErr
	}

	return sent, nil
}

// Отправляет сообщение на указанный email
func (e *EmailService) sendTo(m Emailer, to string) error {
	m.SetTo(to)

	d := gomail.NewDialer(e.d.Host, e.d.Port, e.d.Username, e.d.Password)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m.Message())
}

// Возвращает сообщение для отправки на почту с картинкой вложением
// без заполненного заголовка "To".
// Перед отправкой необходимо заполнить заголовок "To".
func (e *EmailService) availabilityEmailMessage() (Emailer, error) {
	screenshotFullPath := util.GetAbsolutePath(e.d.ScreenshotFilePath)

	emailTemplate, err := e.getEmailTemplate()
	if err != nil {
		return nil, fmt.Errorf("error reading email template: %v", err)
	}
	emailTemplate = fmt.Sprintf(emailTemplate, screenshotCID)

	m := gomail.NewMessage()

	// прикрепляет картинку как вложение
	m.Attach(screenshotFullPath)
	m.SetHeader("From", emailFrom)
	m.SetHeader("Subject", emailSubject)
	m.Embed(screenshotFullPath, gomail.SetHeader(map[string][]string{
		"Content-ID": {"<" + screenshotCID + ">"},
	}))
	m.SetBody("text/html", emailTemplate)

	return &EmailMessage{m}, nil
}

func (e *EmailService) getEmailTemplate() (string, error) {
	path := assetsFolder + emailTemplateFilename
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
