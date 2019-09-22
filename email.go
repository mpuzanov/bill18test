package main

import (
	"log"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/scorredoira/email"
)

//EmailCredentials Структура настройки сервера smtp
type EmailCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Server   string `yaml:"server"`
	Port     int    `yaml:"port"`
}

//SendEmail Отправка почтовых сообщений
func SendEmail(addFrom, addTo, subject, bodyMessage, attachFiles string) error {

	authCreds := EmailCredentials{
		Username: cfg.SettingsSMTP.Username,
		Password: cfg.SettingsSMTP.Password,
		Server:   cfg.SettingsSMTP.Server,
		Port:     cfg.SettingsSMTP.Port,
	}

	// compose the message
	m := email.NewMessage(subject, bodyMessage)
	m.From = mail.Address{Name: addFrom, Address: cfg.SettingsSMTP.Username}
	m.To = []string{addTo}
	m.Subject = subject

	if attachFiles != "" {
		var splitsAttachFiles = strings.Split(attachFiles, ";")
		//log.Printf("%q\n", splitsAttachFiles)
		for _, file := range splitsAttachFiles {
			// add attachments
			if err := m.Attach(file); err != nil {
				return err
			}
		}
	}

	// send it
	auth := smtp.PlainAuth("", cfg.SettingsSMTP.Username, authCreds.Password, authCreds.Server)
	if err := email.Send(authCreds.Server+":25", auth, m); err != nil {
		return err
	}

	log.Println("Email Sent!")
	return nil

}
