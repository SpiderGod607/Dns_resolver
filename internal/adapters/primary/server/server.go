package server

import (
	"dns_resolver/internal/core/resolver"
	"fmt"
	"net"
)

type App struct {
	rootServer []net.IP
	conn       net.PacketConn
	resolver   resolver.DnsResolver
}

func NewApp(
	rootServer []net.IP,
) (*App, error) {

	packetConnection, err := net.ListenPacket("udp", ":53")
	if err != nil {
		return nil, err
	}

	return &App{
		rootServer: rootServer,
		conn:       packetConnection,
		resolver:   resolver.NewDnsResolverImpl(),
	}, nil
}

func (a *App) Run() error {
	fmt.Println("Starting DNS Server..")
	defer a.conn.Close()

	for {
		buf := make([]byte, 512)

		bytesRead, addr, err := a.conn.ReadFrom(buf)

		if err != nil {
			fmt.Printf("Read error from %s: %s \n", addr.String(), err)
			continue
		}

		go a.resolver.HandlePacket(
			a.rootServer,
			a.conn,
			addr,
			buf[:bytesRead],
		)
	}
}
