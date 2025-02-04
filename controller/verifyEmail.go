package controller

import (
	"bytes"
	"html/template"

	"github.com/openziti/zrok/build"
	"github.com/openziti/zrok/controller/emailUi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wneessen/go-mail"
)

type verificationEmail struct {
	EmailAddress string
	VerifyUrl    string
	Version      string
}

func sendVerificationEmail(emailAddress, regToken string) error {
	emailData := &verificationEmail{
		EmailAddress: emailAddress,
		VerifyUrl:    cfg.Registration.RegistrationUrlTemplate + "/" + regToken,
		Version:      build.String(),
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
	if err := msg.From(cfg.Email.From); err != nil {
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
	msg.SetHeader("List-Unsubscribe", "<mailto: invite@zrok.io?subject=unsubscribe>")

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
	t, err := template.ParseFS(emailUi.FS, filename)
	if err != nil {
		return "", errors.Wrapf(err, "error parsing verification email template '%v'", filename)
	}
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, emailData); err != nil {
		return "", errors.Wrapf(err, "error executing verification email template '%v'", filename)
	}
	return buf.String(), nil
}
