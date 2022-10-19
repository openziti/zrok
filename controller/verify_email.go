package controller

import (
	"bytes"
	"github.com/openziti-test-kitchen/zrok/controller/email_ui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wneessen/go-mail"
	"html/template"
)

type verificationEmail struct {
	EmailAddress string
	VerifyUrl    string
}

func sendVerificationEmail(emailAddress, token string) error {
	emailData := &verificationEmail{
		EmailAddress: emailAddress,
		VerifyUrl:    cfg.Registration.RegistrationUrlTemplate + "/" + token,
	}

	plainBody, err := mergeTemplate(emailData, "verify.gotext")
	if err != nil {
		return err
	}
	htmlBody, err := mergeTemplate(emailData, "verify.gohtml")
	if err != nil {
		return err
	}

	msg := mail.NewMsg()
	if err := msg.From("ziggy@zrok.io"); err != nil {
		return errors.Wrap(err, "failed to set from address in verification email")
	}
	if err := msg.To(emailAddress); err != nil {
		return errors.Wrap(err, "failed to sent to address in verification email")
	}
	msg.Subject("Welcome to zrok!")
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
		return errors.Wrap(err, "error creating verification email client")
	}
	if err := client.DialAndSend(msg); err != nil {
		return errors.Wrap(err, "error sending verification email")
	}

	logrus.Infof("verification email sent to '%v'", emailAddress)
	return nil
}

func mergeTemplate(emailData *verificationEmail, filename string) (string, error) {
	t, err := template.ParseFS(email_ui.FS, filename)
	if err != nil {
		return "", errors.Wrapf(err, "error parsing verification email template '%v'", filename)
	}
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, emailData); err != nil {
		return "", errors.Wrapf(err, "error executing verification email template '%v'", filename)
	}
	return buf.String(), nil
}
