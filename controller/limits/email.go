package limits

import (
	"fmt"
	"github.com/openziti/zrok/build"
	"github.com/openziti/zrok/controller/emailUi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wneessen/go-mail"
)

type detailMessage struct {
	lines []string
}

func newDetailMessage() *detailMessage {
	return &detailMessage{}
}

func (m *detailMessage) append(msg string, args ...interface{}) *detailMessage {
	m.lines = append(m.lines, fmt.Sprintf(msg, args...))
	return m
}

func (m *detailMessage) html() string {
	out := ""
	for i := range m.lines {
		out += fmt.Sprintf("<p style=\"text-align: left;\">%s</p>\n", m.lines[i])
	}
	return out
}

func (m *detailMessage) plain() string {
	out := ""
	for i := range m.lines {
		out += fmt.Sprintf("%s\n\n", m.lines[i])
	}
	return out
}

func sendLimitWarningEmail(cfg *emailUi.Config, emailTo string, d *detailMessage) error {
	emailData := &emailUi.WarningEmail{
		EmailAddress: emailTo,
		Version:      build.String(),
	}

	emailData.Detail = d.plain()
	plainBody, err := emailData.MergeTemplate("limitWarning.gotext")
	if err != nil {
		return err
	}

	emailData.Detail = d.html()
	htmlBody, err := emailData.MergeTemplate("limitWarning.gohtml")
	if err != nil {
		return err
	}

	msg := mail.NewMsg()
	if err := msg.From(cfg.From); err != nil {
		return errors.Wrap(err, "failed to set from address in limit warning email")
	}
	if err := msg.To(emailTo); err != nil {
		return errors.Wrap(err, "failed to set to address in limit warning email")
	}

	msg.Subject("zrok Limit Warning Notification")
	msg.SetDate()
	msg.SetMessageID()
	msg.SetBulk()
	msg.SetImportance(mail.ImportanceHigh)
	msg.SetBodyString(mail.TypeTextPlain, plainBody)
	msg.SetBodyString(mail.TypeTextHTML, htmlBody)

	client, err := mail.NewClient(cfg.Host,
		mail.WithPort(cfg.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.Password),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)

	if err != nil {
		return errors.Wrap(err, "error creating limit warning email client")
	}
	if err := client.DialAndSend(msg); err != nil {
		return errors.Wrap(err, "error sending limit warning email")
	}

	logrus.Infof("limit warning email sent to '%v'", emailTo)
	return nil
}
