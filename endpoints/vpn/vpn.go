package vpn

import (
	"fmt"
	"github.com/songgao/water/waterutil"
	"net"
	"strconv"
)

const ZROK_VPN_MTU = 16 * 1024

var (
	zrokIPv4Addr = net.IPv4(10, 'z', 0, 0)
	zrokIPv4     = &net.IPNet{
		IP:   net.IPv4(10, 'z', 0, 0),
		Mask: net.CIDRMask(16, 8*net.IPv4len),
	}

	zrokIPv6Addr = net.IP{0xfd, 0, 'z', 'r', 'o', 'k', 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	zrokIPv6     = &net.IPNet{
		IP: net.IP{0xfd, 0, 'z', 'r', 'o', 'k', // prefix + global ID
			0, 0, // subnet id
			0, 0, 0, 0,
			0, 0, 0, 0,
		},
		Mask: net.CIDRMask(64, 8*net.IPv6len),
	}
)

func DefaultTarget() string {
	l := len(zrokIPv4Addr)
	subnet := net.IPNet{
		IP:   make([]byte, l),
		Mask: zrokIPv4.Mask,
	}

	copy(subnet.IP, zrokIPv4Addr)
	subnet.IP[l-1] = 1
	return subnet.String()
}

type ClientConfig struct {
	Greeting   string
	CIDR       string
	CIDR6      string
	ServerIP   string
	ServerIPv6 string
	Routes     []string
	MTU        int
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
