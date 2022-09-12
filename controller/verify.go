package controller

import (
	"bytes"
	"fmt"
	"github.com/openziti-test-kitchen/zrok/controller/email_ui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/smtp"
)

type verificationEmail struct {
	EmailAddress string
	VerifyUrl    string
}

func sendVerificationEmail(emailAddress, token string, cfg *Config) error {
	t, err := template.ParseFS(email_ui.FS, "verify.gohtml")
	if err != nil {
		return errors.Wrap(err, "error parsing email verification template")
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, &verificationEmail{
		EmailAddress: emailAddress,
		VerifyUrl:    cfg.Registration.RegistrationUrlTemplate + "/" + token,
	})
	if err != nil {
		return errors.Wrap(err, "error executing email verification template")
	}

	subject := "Subject: Welcome to zrok!\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(subject + mime + buf.String())
	auth := smtp.PlainAuth("", cfg.Email.Username, cfg.Email.Password, cfg.Email.Host)
	to := []string{emailAddress}
	err = smtp.SendMail(fmt.Sprintf("%v:%d", cfg.Email.Host, cfg.Email.Port), auth, cfg.Registration.EmailFrom, to, msg)
	if err != nil {
		return errors.Wrap(err, "error sending email verification")
	}

	logrus.Infof("verification email sent to '%v'", emailAddress)
	return nil
}
