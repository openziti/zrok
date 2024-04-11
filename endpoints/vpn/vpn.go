package vpn

import (
	"fmt"
	"github.com/songgao/water/waterutil"
	"net"
	"strconv"
)

const ZROK_VPN_MTU = 16 * 1024

type ClientConfig struct {
	Greeting string
	IP       string
	CIDR     string
	ServerIP string
	Routes   []string
	MTU      int
}

type dest struct {
	addr [4]byte
}

func (d dest) String() string {
	return net.IP(d.addr[:]).String()
}

func (d dest) toInt32() uint32 {
	return uint32(d.addr[0])<<24 + uint32(d.addr[1])<<16 + uint32(d.addr[2])<<8 + uint32(d.addr[3])
}

func ipToDest(addr net.IP) dest {
	d := dest{}
	copy(d.addr[:], addr.To4())
	return d
}

type packet []byte

func (p packet) destination() dest {
	return ipToDest(waterutil.IPv4Destination(p))
}

func (p packet) String() string {
	return fmt.Sprintf("%s %s:%d -> %s:%d %d bytes", p.proto(),
		waterutil.IPv4Source(p), waterutil.IPv4SourcePort(p),
		waterutil.IPv4Destination(p), waterutil.IPv4DestinationPort(p), len(waterutil.IPv4Payload(p)))
}

func (p packet) proto() string {
	proto := waterutil.IPv4Protocol(p)
	switch proto {
	case waterutil.TCP:
		return "tcp"
	case waterutil.UDP:
		return "udp"
	case waterutil.ICMP:
		return "icmp"
	default:
		return strconv.Itoa(int(proto))
	}
}
