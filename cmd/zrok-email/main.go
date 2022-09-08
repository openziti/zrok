package main

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/sirupsen/logrus"
	"net/smtp"
	"os"
)

func init() {
	pfxlog.GlobalInit(logrus.InfoLevel, pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti-test-kitchen/"))
}

func main() {
	from := "ziggy@zrok.io"
	to := []string{"michael@quigley.com"}
	host := "smtp.email.us-ashburn-1.oci.oraclecloud.com"
	port := "587"

	msg := "Subject: Ziggy\r\n" +
		"\r\n" +
		"Hello from Ziggy!\r\n"
	body := []byte(msg)

	auth := smtp.PlainAuth("", os.Args[1], os.Args[2], host)

	err := smtp.SendMail(host+":"+port, auth, from, to, body)
	if err != nil {
		panic(err)
	}

	logrus.Infof("message sent")
}
