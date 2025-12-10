package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
)

const CTRL_PUBLISH byte = 2

func main() {
	brokerAddr := flag.String("broker", "localhost:8080", "Broker address")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Usage: publisher -broker host:port <topic> <message>\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	topic := flag.Arg(0)
	message := flag.Arg(1)

	conn, err := net.Dial("tcp", *brokerAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to broker on %s: %v\n", *brokerAddr, err)
		os.Exit(1)
	}
	defer conn.Close()

	topicByte := []byte(topic)
	msgByte := []byte(message)

	packet := []byte{CTRL_PUBLISH}
	packet = append(packet, 0, 0)
	packet = append(packet, topicByte...)
	binary.BigEndian.PutUint16(packet[1:3], uint16(len(topicByte)))

	payloadLen := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadLen, uint32(len(msgByte)))
	packet = append(packet, payloadLen...)
	packet = append(packet, msgByte...)

	if _, err := conn.Write(packet); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Published: \"%s\" -> topic \"%s\" via %s\n", message, topic, *brokerAddr)
}
