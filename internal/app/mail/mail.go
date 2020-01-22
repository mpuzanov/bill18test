package mail

import (
	"log"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/mpuzanov/bill18test/internal/app/config"
	"github.com/scorredoira/email"
)

//SendEmail Отправка почтовых сообщений
func SendEmail(cfg *config.Config, addFrom, addTo, subject, bodyMessage, attachFiles string) error {

	authCreds := config.EmailCredentials{
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
	var auth smtp.Auth
	if authCreds.Password != "" {
		auth = smtp.PlainAuth("", cfg.SettingsSMTP.Username, authCreds.Password, authCreds.Server)
	}
	if err := email.Send(authCreds.Server+":"+authCreds.Port, auth, m); err != nil {
		return err
	}

	log.Println("Email Sent!")
	return nil

}
