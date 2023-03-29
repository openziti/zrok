package limits

import (
	"github.com/openziti/zrok/build"
	"github.com/openziti/zrok/controller/emailUi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wneessen/go-mail"
)

func sendLimitWarningEmail(cfg *emailUi.Config, emailTo, detail string) error {
	emailData := &emailUi.WarningEmail{
		EmailAddress: emailTo,
		Detail:       detail,
		Version:      build.String(),
	}

	plainBody, err := emailData.MergeTemplate("limitWarning.gotext")
	if err != nil {
		return err
	}
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

	msg.Subject("Limit Warning Notification")
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
