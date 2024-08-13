package resolver

import (
	"fmt"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

func (r *DnsResolverImpl) PerformDnsQuery(
	rootServers []net.IP,
	question dnsmessage.Question,
) (*dnsmessage.Message, error) {

	servers := rootServers
	for i := 0; i < 3; i++ {
		responseParser, header, err := r.SendDnsPacket(servers, question)
		if err != nil {
			return nil, fmt.Errorf("Error sending dns packet %s", err)
		}

		parsedAnswers, err := responseParser.AllAnswers()
		if err != nil {
			return nil, fmt.Errorf("Error passing dns answers %s", err)
		}

		if header.Authoritative {
			fmt.Println("got authoritative answers")
			return &dnsmessage.Message{
				Header:  dnsmessage.Header{Response: true},
				Answers: parsedAnswers,
			}, nil
		}

		nextServers, err := getNextServerToHitInHierarcy(r, responseParser, &rootServers)
		if err != nil {
			return &dnsmessage.Message{
				Header: dnsmessage.Header{RCode: dnsmessage.RCodeServerFailure},
			}, nil
		}

		servers = nextServers
	}

	return &dnsmessage.Message{
		Header: dnsmessage.Header{RCode: dnsmessage.RCodeServerFailure},
	}, nil
}

func getNextServerToHitInHierarcy(r DnsResolver, parser *dnsmessage.Parser, rootServers *[]net.IP) ([]net.IP, error) {
	parsedAuthorities, err := parser.AllAuthorities()
	if err != nil {
		return nil, fmt.Errorf("Failed to pasedAuthorities %s", err)
	}
	if len(parsedAuthorities) == 0 {
		return nil, fmt.Errorf("no authorites found")
	}

	nextServerDomains := make([]string, len(parsedAuthorities))
	for i, authority := range parsedAuthorities {
		if authority.Header.Type == dnsmessage.TypeNS {
			nextServerDomains[i] = authority.Body.(*dnsmessage.NSResource).NS.String()
		}
	}

	additionals, err := parser.AllAdditionals()
	if err != nil {
		return nil, fmt.Errorf("Failed to additionals %s", err)
	}

	fmt.Printf("got name answers %s \n", nextServerDomains)

	nextSever := []net.IP{}
	for _, addditional := range additionals {
		if addditional.Header.Type == dnsmessage.TypeA {
			for _, nameServer := range nextServerDomains {
				if addditional.Header.Name.String() == nameServer {
					ip := addditional.Body.(*dnsmessage.AResource).A[:]
					nextSever = append(nextSever, ip)
				}
			}
		}
	}

	if len(nextSever) == 0 {
		for _, nameServer := range nextServerDomains {
			response, err := r.PerformDnsQuery(*rootServers,
				dnsmessage.Question{
					Name:  dnsmessage.MustNewName(nameServer),
					Type:  dnsmessage.TypeA,
					Class: dnsmessage.ClassINET,
				})

			if err != nil {
				fmt.Printf("lookup of nameserver %s failed: %err\n", nameServer, err)
			} else {
				for _, answer := range response.Answers {
					if answer.Header.Type == dnsmessage.TypeA {
						nextSever = append(nextSever, answer.Body.(*dnsmessage.AResource).A[:])
						return nextSever, nil
					}
				}
			}

		}
	}

	return nextSever, nil
}
