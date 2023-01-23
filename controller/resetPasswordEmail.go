package controller

import (
	"bytes"
	"fmt"
	"github.com/openziti/zrok/build"
	"html/template"

	"github.com/openziti/zrok/controller/emailUi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wneessen/go-mail"
)

type resetPasswordEmail struct {
	EmailAddress string
	Url          string
	Version      string
}

func sendResetPasswordEmail(emailAddress, token string) error {
	emailData := &resetPasswordEmail{
		EmailAddress: emailAddress,
		Url:          fmt.Sprintf("%s/%s", cfg.ResetPassword.ResetUrlTemplate, token),
		Version:      build.String(),
	}

	plainBody, err := emailData.mergeTemplate("resetPassword.gotext")
	if err != nil {
		return err
	}
	htmlBody, err := emailData.mergeTemplate("resetPassword.gohtml")
	if err != nil {
		return err
	}

	msg := mail.NewMsg()
	if err := msg.From(cfg.Email.From); err != nil {
		return errors.Wrap(err, "failed to set from address in reset password email")
	}
	if err := msg.To(emailAddress); err != nil {
		return errors.Wrap(err, "failed to set to address in reset password email")
	}

	msg.Subject("Password Reset Request")
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
		return errors.Wrap(err, "error creating reset password email client")
	}
	if err := client.DialAndSend(msg); err != nil {
		return errors.Wrap(err, "error sending reset password email")
	}

	logrus.Infof("reset password email sent to '%v'", emailAddress)
	return nil
}

func (fpe resetPasswordEmail) mergeTemplate(filename string) (string, error) {
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
