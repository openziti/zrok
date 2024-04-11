package vpn

import (
	"encoding/json"
	"fmt"
	_ "github.com/net-byte/vtun/tun"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/endpoints"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/songgao/water/waterutil"
	"net"
	"time"
)

type BackendConfig struct {
	IdentityPath    string
	EndpointAddress string
	ShrToken        string
	RequestsChan    chan *endpoints.Request
}

type Backend struct {
	cfg      *BackendConfig
	listener edge.Listener

	cidr net.IPAddr
}

func NewBackend(cfg *BackendConfig) (*Backend, error) {

	options := ziti.ListenOptions{
		ConnectTimeout:               5 * time.Minute,
		WaitForNEstablishedListeners: 1,
	}
	zcfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	zctx, err := ziti.NewContext(zcfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}
	listener, err := zctx.ListenWithOptions(cfg.ShrToken, &options)
	if err != nil {
		return nil, errors.Wrap(err, "error listening")
	}
	b := &Backend{
		cfg:      cfg,
		listener: listener,
	}
	return b, nil
}

func (b *Backend) Run() error {
	logrus.Info("started")
	defer logrus.Info("exited")

	for {
		if conn, err := b.listener.Accept(); err == nil {
			go b.handle(conn)
		} else {
			return err
		}
	}
}

type ipProto waterutil.IPProtocol

func (p ipProto) String() string {
	switch p {
	case waterutil.TCP:
		return "tcp"
	case waterutil.UDP:
		return "udp"
	case waterutil.ICMP:
		return "icmp"
	default:
		return fmt.Sprintf("proto[%d]", p)
	}
}

func (b *Backend) handle(conn net.Conn) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	cfg := &ClientConfig{
		Greeting: "Welcome to zrok VPN",
	}
	j, err := json.Marshal(&cfg)
	if err != nil {
		return
	}
	conn.Write(j)

	buf := make([]byte, 16*1024)
	for {
		read, err := conn.Read(buf)
		if err != nil {
			logrus.Error("read error", err)
			return
		}
		pkt := buf[:read]
		logrus.Infof("read packet %d bytes %v %v:%v -> %v:%v", read,
			ipProto(waterutil.IPv4Protocol(pkt)),
			waterutil.IPv4Source(pkt), waterutil.IPv4SourcePort(pkt),
			waterutil.IPv4Destination(pkt), waterutil.IPv4DestinationPort(pkt))

	}
}
