package main

import (
	"dns_resolver/internal/adapters/primary/server"
	"log"
	"net"
	"strings"
)

func main() {
	app, err := server.NewApp(
		getRootServers(),
	)
	if err != nil {
		log.Fatal(err)
	}

	app.Run()
}

func getRootServers() []net.IP {

	servers := "198.41.0.4,199.9.14.201,192.33.4.12,199.7.91.13,192.203.230.10,192.5.5.241,192.112.36.4,198.97.190.53"

	rootServers := []net.IP{}
	for _, rootServer := range strings.Split(servers, ",") {
		rootServers = append(rootServers, net.ParseIP(rootServer))
	}
	return rootServers

}
