package controller

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/openziti/channel/v2"
	"github.com/openziti/channel/v2/websockets"
	"github.com/openziti/edge/rest_util"
	"github.com/openziti/fabric/event"
	"github.com/openziti/fabric/pb/mgmt_pb"
	"github.com/openziti/identity"
	"github.com/openziti/sdk-golang/ziti/constants"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"time"
)

func RunUsageAgent(cfg *Config) error {
	zitiApiEndpointUrl, err := url.Parse(cfg.Ziti.ApiEndpoint)
	if err != nil {
		return err
	}

	wsUrl := "wss://" + zitiApiEndpointUrl.Host + "/fabric/v1/ws-api"
	logrus.Infof("wsUrl: %v", wsUrl)

	caCerts, err := rest_util.GetControllerWellKnownCas(cfg.Ziti.ApiEndpoint)
	if err != nil {
		return err
	}
	caPool := x509.NewCertPool()
	for _, ca := range caCerts {
		caPool.AddCert(ca)
	}

	authenticator := rest_util.NewAuthenticatorUpdb(cfg.Ziti.Username, cfg.Ziti.Password)
	authenticator.RootCas = caPool

	apiSession, err := authenticator.Authenticate(zitiApiEndpointUrl)
	if err != nil {
		return err
	}

	dialer := &websocket.Dialer{
		TLSClientConfig: &tls.Config{
			RootCAs: caPool,
		},
		HandshakeTimeout: 5 * time.Second,
	}

	conn, resp, err := dialer.Dial(wsUrl, http.Header{constants.ZitiSession: []string{*apiSession.Token}})
	if err != nil {
		if resp != nil {
			if body, rerr := io.ReadAll(resp.Body); rerr == nil {
				logrus.Errorf("response body '%v': %v", string(body), err)
			}
		} else {
			logrus.Errorf("no response from websocket dial: %v", err)
		}
	}

	id := &identity.TokenId{Token: "mgmt"}
	underlayFactory := websockets.NewUnderlayFactory(id, conn, nil)

	closeNotify := make(chan struct{})

	bindHandler := func(binding channel.Binding) error {
		binding.AddReceiveHandler(int32(mgmt_pb.ContentType_StreamEventsEventType), &usageAgent{})
		binding.AddCloseHandler(channel.CloseHandlerF(func(ch channel.Channel) {
			close(closeNotify)
		}))
		return nil
	}

	ch, err := channel.NewChannel("mgmt", underlayFactory, channel.BindHandlerF(bindHandler), nil)
	if err != nil {
		return err
	}

	streamEventsRequest := map[string]interface{}{}
	streamEventsRequest["format"] = "json"
	streamEventsRequest["subscriptions"] = []*event.Subscription{
		{
			Type: "fabric.usage",
			Options: map[string]interface{}{
				"version": uint8(3),
			},
		},
	}

	msgBytes, err := json.Marshal(streamEventsRequest)
	if err != nil {
		return err
	}

	requestMsg := channel.NewMessage(int32(mgmt_pb.ContentType_StreamEventsRequestType), msgBytes)
	responseMsg, err := requestMsg.WithTimeout(5 * time.Second).SendForReply(ch)
	if err != nil {
		return err
	}

	if responseMsg.ContentType == channel.ContentTypeResultType {
		result := channel.UnmarshalResult(responseMsg)
		if result.Success {
			logrus.Infof("event stream started: %v", result.Message)
		} else {
			return errors.Wrap(err, "error starting event streaming")
		}
	} else {
		return errors.Errorf("unexpected response type %v", responseMsg.ContentType)
	}

	<-closeNotify
	return nil
}

type usageAgent struct{}

func (ua *usageAgent) HandleReceive(msg *channel.Message, _ channel.Channel) {
	decoder := json.NewDecoder(bytes.NewReader(msg.Body))
	for {
		event := make(map[string]interface{})
		err := decoder.Decode(&event)
		if err == io.EOF {
			break
		}
		if err == nil {
			ui := &metrics.UsageIngester{}
			if err := ui.Ingest(event); err != nil {
				logrus.Errorf("error ingesting '%v': %v", string(msg.Body), err)
			}
		} else {
			logrus.Errorf("error parsing '%v': %v", string(msg.Body), err)
		}
	}
}
