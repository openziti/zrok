package vpn

import (
	"encoding/json"
	"io"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/google/go-cmp/cmp"
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
)

type BackendConfig struct {
	IdentityPath    string
	EndpointAddress string
	ShrToken        string
	RequestsChan    chan *endpoints.Request
	SuperNetwork    bool
}

type client struct {
	conn net.Conn
}

type Backend struct {
	cfg      *BackendConfig
	listener edge.Listener

	addr    net.IP
	addr6   net.IP
	subnet  *net.IPNet
	subnet6 *net.IPNet
	tun     *water.Interface
	mtu     int

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
	if cfg.SuperNetwork {
		zcfg.MaxDefaultConnections = 2
		zcfg.MaxControlConnections = 1
		logrus.Warnf("super networking enabled")
	}
	zctx, err := ziti.NewContext(zcfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}
	listener, err := zctx.ListenWithOptions(cfg.ShrToken, &options)
	if err != nil {
		return nil, errors.Wrap(err, "error listening")
	}

	addr6 := zrokIPv6Addr
	addr4 := zrokIPv4Addr
	sub4 := zrokIPv4
	sub6 := zrokIPv6

	if cfg.EndpointAddress != "" {
		addr4, sub4, err = net.ParseCIDR(cfg.EndpointAddress)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse VPN subnet config")
		}
	}

	b := &Backend{
		cfg:      cfg,
		listener: listener,
		mtu:      ZROK_VPN_MTU,
		clients: cmap.NewWithCustomShardingFunction[dest, *client](func(key dest) uint32 {
			return key.toInt32()
		}),
		addr:    addr4,
		addr6:   addr6,
		subnet:  sub4,
		subnet6: sub6,
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
				b.cfg.RequestsChan <- &endpoints.Request{
					Stamp:      time.Now(),
					RemoteAddr: dest.String(),
					Method:     "DISCONNECTED",
				}

				logrus.WithError(err).Errorf("failed to write packet to clt[%v]", dest)
				_ = clt.conn.Close()
				b.clients.Remove(dest)
			}
		} else {
			if b.subnet.Contains(net.IP(dest.addr[:])) {
				logrus.Errorf("no client with address[%v]", dest)
			}
		}
	}
}

func (b *Backend) Run() error {
	logrus.Info("started")
	defer logrus.Info("exited")

	bits, _ := b.subnet.Mask.Size()
	bits6, _ := b.subnet6.Mask.Size()

	tunCfg := config.Config{
		ServerIP:   b.addr.String(),
		ServerIPv6: b.addr6.String(),
		CIDR:       b.addr.String() + "/" + strconv.Itoa(bits),
		CIDRv6:     b.addr6.String() + "/" + strconv.Itoa(bits6),
		MTU:        ZROK_VPN_MTU,
		Verbose:    true,
	}
	logrus.Infof("%+v", tunCfg)
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

	ipv4, ipv6 := b.nextIP()
	ip := ipToDest(ipv4)

	bits, _ := b.subnet.Mask.Size()
	bits6, _ := b.subnet6.Mask.Size()

	cfg := &ClientConfig{
		Greeting:   "Welcome to zrok VPN",
		ServerIP:   b.addr.String(),
		ServerIPv6: b.addr6.String(),
		CIDR:       ipv4.String() + "/" + strconv.Itoa(bits),
		CIDR6:      ipv6.String() + "/" + strconv.Itoa(bits6),
		MTU:        b.mtu,
	}

	b.cfg.RequestsChan <- &endpoints.Request{
		Stamp:      time.Now(),
		RemoteAddr: ipv4.String(),
		Method:     "CONNECTED",
		Path:       cfg.ServerIP,
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
			if err != io.EOF {
				logrus.WithError(err).Error("read error")
			}
			b.cfg.RequestsChan <- &endpoints.Request{
				Stamp:      time.Now(),
				RemoteAddr: ipv4.String(),
				Method:     "DISCONNECTED",
			}
			return
		}
		pkt := packet(buf[:read])
		logrus.WithField("packet", pkt).Trace("read from ziti")
		_, err = b.tun.Write(pkt)
		if err != nil {
			logrus.WithError(err).Error("failed to write packet to tun")
			return
		}
	}
}

func (b *Backend) nextIP() (net.IP, net.IP) {
	ip4 := make([]byte, len(b.subnet.IP))
	for {
		copy(ip4, b.subnet.IP)
		n := b.counter.Add(1)
		if n == 0 {
			continue
		}

		for i := 0; i < len(ip4); i++ {
			b := (n >> (i * 8)) % 0xff
			ip4[len(ip4)-1-i] ^= byte(b)
		}

		// subnet overflow
		if !b.subnet.Contains(ip4) {
			b.counter.Store(1)
			continue
		}

		if cmp.Equal(b.addr, ip4) {
			continue
		}

		if b.clients.Has(ipToDest(ip4)) {
			continue
		}

		break
	}

	ip6 := append([]byte{}, b.subnet6.IP...)
	copy(ip6[net.IPv6len-net.IPv4len:], ip4)

	return ip4, ip6
}
