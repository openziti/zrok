package controller

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/openziti/channel/v2"
	"github.com/openziti/channel/v2/websockets"
	"github.com/openziti/edge/rest_util"
	"github.com/openziti/fabric/event"
	"github.com/openziti/fabric/pb/mgmt_pb"
	"github.com/openziti/identity"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type usageAgent struct{}

func RunUsageAgent(cfg *Config) error {
	wsUrl := "wss://127.0.0.1:1280/ws-api"

	caCerts, err := rest_util.GetControllerWellKnownCas(cfg.Ziti.ApiEndpoint)
	if err != nil {
		return err
	}
	caPool := x509.NewCertPool()
	for _, ca := range caCerts {
		caPool.AddCert(ca)
	}

	dialer := &websocket.Dialer{
		TLSClientConfig: &tls.Config{
			RootCAs: caPool,
		},
		HandshakeTimeout: 5 * time.Second,
	}

	conn, resp, err := dialer.Dial(wsUrl, http.Header{})
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
		&event.Subscription{
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

func (ua *usageAgent) HandleReceive(msg *channel.Message, _ channel.Channel) {
	fmt.Println(string(msg.Body))
}
