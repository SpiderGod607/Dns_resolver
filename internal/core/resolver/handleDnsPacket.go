package resolver

import (
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

func (r *DnsResolverImpl) HandlePacket(
	rootServers []net.IP,
	conn net.PacketConn,
	addr net.Addr,
	request []byte,
) error {
	p := dnsmessage.Parser{}
	header, err := p.Start(request)
	if err != nil {
		return err
	}

	question, err := p.Question()
	if err != nil {
		return err
	}

	response, err := r.PerformDnsQuery(rootServers, question)
	if err != nil {
		return err
	}
	response.Header.ID = header.ID

	responseInBytes, err := response.Pack()
	if err != nil {
		return err
	}

	_, err = conn.WriteTo(responseInBytes, addr)
	if err != nil {
		return err
	}
	return nil
}
