package resolver

import (
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

type DnsResolver interface {
	HandlePacket(
		rootServers []net.IP,
		pc net.PacketConn, addr net.Addr, buf []byte,
	) error

	SendDnsPacket(
		serversToSendPacketTo []net.IP,
		questions dnsmessage.Question,
	) (*dnsmessage.Parser, *dnsmessage.Header, error)

	PerformDnsQuery(
		rootServers []net.IP,
		question dnsmessage.Question,
	) (*dnsmessage.Message, error)
}

type DnsResolverImpl struct{}

func NewDnsResolverImpl() *DnsResolverImpl {
	return &DnsResolverImpl{}
}
