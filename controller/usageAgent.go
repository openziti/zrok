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
	"github.com/openziti/zrok/util"
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
	//logrus.Infof(string(msg.Body))
	decoder := json.NewDecoder(bytes.NewReader(msg.Body))
	for {
		event := make(map[string]interface{})
		err := decoder.Decode(&event)
		if err == io.EOF {
			break
		}
		if err == nil {
			if ns, found := event["namespace"]; found && ns == "fabric.usage" {
				start := float64(0)
				if v, found := event["interval_start_utc"]; found {
					if vFloat64, ok := v.(float64); ok {
						start = vFloat64
					} else {
						logrus.Error("unable to assert 'interval_start_utc'")
					}
				} else {
					logrus.Error("missing 'interval_start_utc'")
				}
				clientId := ""
				serviceId := ""
				if v, found := event["tags"]; found {
					if tags, ok := v.(map[string]interface{}); ok {
						if v, found := tags["clientId"]; found {
							if vStr, ok := v.(string); ok {
								clientId = vStr
							} else {
								logrus.Error("unable to assert 'tags/clientId'")
							}
						} else {
							logrus.Errorf("missing 'tags/clientId'")
						}
						if v, found := tags["serviceId"]; found {
							if vStr, ok := v.(string); ok {
								serviceId = vStr
							} else {
								logrus.Error("unable to assert 'tags/serviceId'")
							}
						} else {
							logrus.Error("missing 'tags/serviceId'")
						}
					} else {
						logrus.Errorf("unable to assert 'tags'")
					}
				} else {
					logrus.Errorf("missing 'tags'")
				}
				tx := int64(0)
				rx := int64(0)
				if v, found := event["usage"]; found {
					if usage, ok := v.(map[string]interface{}); ok {
						if v, found := usage["egress.tx"]; found {
							if vFloat64, ok := v.(float64); ok {
								tx = int64(vFloat64)
							} else {
								logrus.Error("unable to assert 'usage/egress.tx'")
							}
						} else {
							logrus.Error("missing 'usage/egress.tx'")
						}
						if v, found := usage["egress.rx"]; found {
							if vFloat64, ok := v.(float64); ok {
								rx = int64(vFloat64)
							} else {
								logrus.Error("unable to assert 'usage/egress.rx'")
							}
						} else {
							logrus.Error("missing 'usage/egress.rx'")
						}
					} else {
						logrus.Error("unabel to assert 'usage'")
					}
				} else {
					logrus.Error("missing 'usage'")
				}
				circuitId := ""
				if v, found := event["circuit_id"]; found {
					if vStr, ok := v.(string); ok {
						circuitId = vStr
					} else {
						logrus.Error("unable to assert 'circuit_id'")
					}
				} else {
					logrus.Error("missing 'circuit_id'")
				}

				logrus.Infof("usage: start '%d', serviceId '%v', clientId '%v', circuitId '%v' [rx: %v, tx: %v]", int64(start), serviceId, clientId, circuitId, util.BytesToSize(rx), util.BytesToSize(tx))

			} else {
				logrus.Errorf("not 'fabric.usage'")
			}
		} else {
			logrus.Errorf("error parsing '%v': %v", string(msg.Body), err)
		}
	}
}
