package lib

import (
	"bytes"
	"text/template"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer *gomail.Dialer
}

func NewMailer() *Mailer {

	/// TODO - move these to env variables
	dialer := gomail.NewDialer("smtp.gmail.com", 587, "jdrew153@gmail.com", "fvxctgqtewjfupfw")
	return &Mailer{
		dialer: dialer,
	}
}


func (m *Mailer) SendMail(to string, subject string, username string) error {

	t, err := template.ParseFiles("../templates/welcome.html")
	if err != nil {
		panic(err)
	}

	var buffer bytes.Buffer

	t.Execute(&buffer, struct {
		Username string
	} {
		Username: username,
	})

	body := buffer.String()

	mail := gomail.NewMessage()
	mail.SetHeader("From", "jdrew153@gmail.com")
	mail.SetHeader("To", to)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/html", body)


	if err := m.dialer.DialAndSend(mail); err != nil {
		panic(err)
	}

	return nil
}

