package resolver

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

func (r *DnsResolverImpl) SendDnsPacket(
	serversToSendPacketTo []net.IP,
	questions dnsmessage.Question,
) (*dnsmessage.Parser, *dnsmessage.Header, error) {

	fmt.Printf("Sending dns questions for %s to servers %+v \n", questions.Name.String(), serversToSendPacketTo)
	id, error := generateRandomID()
	if error != nil {
		return nil, nil, fmt.Errorf("Failed to generateRandomID %s", error)
	}

	message := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:       id,
			Response: false,
			OpCode:   dnsmessage.OpCode(0),
		},
		Questions: []dnsmessage.Question{questions},
	}
	return sendPacketToServer(&message, serversToSendPacketTo)
}

func sendPacketToServer(
	message *dnsmessage.Message,
	servers []net.IP,
) (*dnsmessage.Parser, *dnsmessage.Header, error) {

	messageInBytes, err := message.Pack()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to pack message %s", err)
	}

	var conn net.Conn
	for _, server := range servers {
		fmt.Println("Hitting server " + server.String() + ":53")
		conn, err = net.Dial("udp", server.String()+":53")
		if err == nil {
			break
		}
	}

	if conn == nil {
		return nil, nil, fmt.Errorf("Failed to connect to server %s", err)
	}

	_, err = conn.Write(messageInBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed write to buff %s", err)
	}

	answer := make([]byte, 512)
	endOfAnswer, err := bufio.NewReader(conn).Read(answer)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed conver buff to bytes %s", err)
	}
	conn.Close()

	var p dnsmessage.Parser
	header, err := p.Start(answer[:endOfAnswer])
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to parse header %s", err)
	}

	if ok, err := isAllQuestionsAnswered(&p, message); !ok {
		return nil, nil, err
	}

	err = p.SkipAllQuestions()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed skip questions: %s", err)
	}

	return &p, &header, nil
}

func isAllQuestionsAnswered(p *dnsmessage.Parser, message *dnsmessage.Message) (bool, error) {
	questions, err := p.AllQuestions()
	if err != nil {
		return false, fmt.Errorf("failed to get quetions %s", err)
	}
	if len(questions) != len(message.Questions) {
		return false, fmt.Errorf("answer packet doesn't have the same amount of questions")
	}
	return true, nil
}

func generateRandomID() (uint16, error) {
	max := ^uint16(0)
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, fmt.Errorf("Failed to random numer %s", err)
	}
	return uint16(randomNumber.Int64()), nil
}
