package metrics

import (
	"fmt"
	"time"

	"github.com/openziti/zrok/v2/util"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Usage struct {
	ProcessedStamp time.Time
	IntervalStart  time.Time
	ZitiServiceId  string
	ZitiCircuitId  string
	ShareToken     string
	EnvironmentId  int64
	AccountId      int64
	FrontendTx     int64
	FrontendRx     int64
	BackendTx      int64
	BackendRx      int64
}

func (u Usage) String() string {
	out := "Usage {"
	out += fmt.Sprintf("processed '%v'", u.ProcessedStamp)
	out += ", " + fmt.Sprintf("interval '%v'", u.IntervalStart)
	out += ", " + fmt.Sprintf("service '%v'", u.ZitiServiceId)
	out += ", " + fmt.Sprintf("circuit '%v'", u.ZitiCircuitId)
	out += ", " + fmt.Sprintf("share '%v'", u.ShareToken)
	out += ", " + fmt.Sprintf("environment '%d'", u.EnvironmentId)
	out += ", " + fmt.Sprintf("account '%v'", u.AccountId)
	out += ", " + fmt.Sprintf("fe {rx %v, tx %v}", util.BytesToSize(u.FrontendRx), util.BytesToSize(u.FrontendTx))
	out += ", " + fmt.Sprintf("be {rx %v, tx %v}", util.BytesToSize(u.BackendRx), util.BytesToSize(u.BackendTx))
	out += "}"
	return out
}

type UsageSink interface {
	Handle(u *Usage) error
}

type ZitiEventJson string

type ZitiEventJsonMsg struct {
	data ZitiEventJson
}

func (e *ZitiEventJsonMsg) Data() ZitiEventJson {
	return e.data
}

func (e *ZitiEventJsonMsg) Ack() error {
	return nil
}

type ZitiEventAMQP struct {
	data ZitiEventJson
	msg  amqp.Delivery
}

func (e *ZitiEventAMQP) Data() ZitiEventJson {
	return e.data
}

func (e *ZitiEventAMQP) Ack() error {
	return e.msg.Ack(false)
}

type ZitiEventMsg interface {
	Data() ZitiEventJson
	Ack() error
}

type ZitiEventJsonSource interface {
	Start(chan ZitiEventMsg) (join chan struct{}, err error)
	Stop()
}

type ZitiEventJsonSink interface {
	Handle(event ZitiEventJson) error
}
