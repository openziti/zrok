package controller

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/openziti/zrok/controller/emailUi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wneessen/go-mail"
)

type forgotPasswordEmail struct {
	EmailAddress      string
	ForgotPasswordUrl string
}

func sendForgotPasswordEmail(emailAddress, token string) error {
	emailData := &forgotPasswordEmail{
		EmailAddress:      emailAddress,
		ForgotPasswordUrl: fmt.Sprintf("%s?token=%s", cfg.Account.ForgotPasswordUrlTemplate, token),
	}

	plainBody, err := emailData.mergeTemplate("forgotPassword.gotext")
	if err != nil {
		return err
	}
	htmlBody, err := emailData.mergeTemplate("forgotPassword.gohtml")
	if err != nil {
		return err
	}

	msg := mail.NewMsg()
	if err := msg.From(cfg.Registration.EmailFrom); err != nil {
		return errors.Wrap(err, "failed to set from address in forgot password email")
	}
	if err := msg.To(emailAddress); err != nil {
		return errors.Wrap(err, "failed to set to address in forgot password email")
	}

	msg.Subject("zrok Forgot Password")
	msg.SetDate()
	msg.SetMessageID()
	msg.SetBulk()
	msg.SetImportance(mail.ImportanceHigh)
	msg.SetBodyString(mail.TypeTextPlain, plainBody)
	msg.SetBodyString(mail.TypeTextHTML, htmlBody)

	client, err := mail.NewClient(cfg.Email.Host,
		mail.WithPort(cfg.Email.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(cfg.Email.Username),
		mail.WithPassword(cfg.Email.Password),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)

	if err != nil {
		return errors.Wrap(err, "error creating forgot password email client")
	}
	if err := client.DialAndSend(msg); err != nil {
		return errors.Wrap(err, "error sending forgot password email")
	}

	logrus.Infof("forgot password email sent to '%v'", emailAddress)
	return nil
}

func (fpe forgotPasswordEmail) mergeTemplate(filename string) (string, error) {
	t, err := template.ParseFS(emailUi.FS, filename)
	if err != nil {
		return "", errors.Wrapf(err, "error parsing verification email template '%v'", filename)
	}
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, fpe); err != nil {
		return "", errors.Wrapf(err, "error executing verification email template '%v'", filename)
	}
	return buf.String(), nil
}
