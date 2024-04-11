package vpn

import (
	"encoding/json"
	"github.com/net-byte/vtun/common/config"
	"github.com/net-byte/vtun/tun"
	_ "github.com/net-byte/vtun/tun"
	"github.com/net-byte/water"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/endpoints"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/songgao/water/waterutil"
	"net"
	"sync/atomic"
	"time"
)

type BackendConfig struct {
	IdentityPath    string
	EndpointAddress string
	ShrToken        string
	RequestsChan    chan *endpoints.Request
}

type client struct {
	conn net.Conn
}

type Backend struct {
	cfg      *BackendConfig
	listener edge.Listener

	cidr net.IPAddr
	tun  *water.Interface
	mtu  int

	counter atomic.Uint32
	clients cmap.ConcurrentMap[dest, *client]
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
		mtu:      ZROK_VPN_MTU,
		clients: cmap.NewWithCustomShardingFunction[dest, *client](func(key dest) uint32 {
			return key.toInt32()
		}),
	}
	b.counter.Store(1)
	return b, nil
}

func (b *Backend) readTun() {
	buf := make([]byte, ZROK_VPN_MTU)
	for {
		n, err := b.tun.Read(buf)
		if err != nil {
			logrus.WithError(err).Error("failed to read tun device")
			// handle? error
			panic(err)
			return
		}
		pkt := packet(buf[:n])
		if !waterutil.IsIPv4(pkt) {
			continue
		}

		logrus.WithField("packet", pkt).Trace("read from tun device")
		dest := pkt.destination()

		if clt, ok := b.clients.Get(dest); ok {
			_, err := clt.conn.Write(pkt)
			if err != nil {
				logrus.WithError(err).Errorf("failed to write packet to clt[%v]", dest)
				_ = clt.conn.Close()
				b.clients.Remove(dest)
			}
		} else {
			logrus.Errorf("no client with address[%v]", dest)
		}
	}
}

func (b *Backend) Run() error {
	logrus.Info("started")
	defer logrus.Info("exited")

	tunCfg := config.Config{
		ServerIP:   "192.168.127.1",
		ServerIPv6: "fced::ffff:c0a8:7f01",
		CIDR:       "192.168.127.1/24",
		CIDRv6:     "fced::ffff:c0a8:7f01/64",
		MTU:        ZROK_VPN_MTU,
		Verbose:    true,
	}

	b.tun = tun.CreateTun(tunCfg)
	defer func() {
		_ = b.tun.Close()
	}()

	go b.readTun()

	for {
		if conn, err := b.listener.Accept(); err == nil {
			go b.handle(conn)
		} else {
			return err
		}
	}
}

func (b *Backend) handle(conn net.Conn) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	num := uint32(0)
	for num == 0 || num == 1 {
		num = b.counter.Add(1)
		num = num % 256
	}

	ipv4 := net.IPv4(192, 168, 127, byte(num))
	ip := ipToDest(ipv4)

	cfg := &ClientConfig{
		Greeting: "Welcome to zrok VPN",
		IP:       ipv4.String(),
		ServerIP: "192.168.127.1",
		CIDR:     ipv4.String() + "/24",
		MTU:      b.mtu,
	}

	j, err := json.Marshal(&cfg)
	if err != nil {
		logrus.WithError(err).Error("failed to write client VPN config")
		return
	}
	_, err = conn.Write(j)
	if err != nil {
		logrus.WithError(err).Error("failed to write client VPN config")
		return
	}

	clt := &client{conn: conn}
	b.clients.Set(ip, clt)

	buf := make([]byte, b.mtu)
	for {
		read, err := conn.Read(buf)
		if err != nil {
			logrus.Error("read error", err)
			return
		}
		pkt := packet(buf[:read])
		logrus.WithField("packet", pkt).Info("read from ziti")
		_, err = b.tun.Write(pkt)
		if err != nil {
			logrus.WithError(err).Error("failed to write packet to tun")
		}
	}
}
