package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
)

const controlSubscribe byte = 1

func main() {
	brokerAddr := flag.String("broker", "localhost:8080", "Broker address (host:port)")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: subscriber -broker host:port <topic>\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	topic := flag.Arg(0)

	conn, err := net.Dial("tcp", *brokerAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to broker %s: %v\n", *brokerAddr, err)
		os.Exit(1)
	}
	defer conn.Close()

	topicB := []byte(topic)
	packet := []byte{controlSubscribe}
	packet = append(packet, 0, 0)
	packet = append(packet, topicB...)
	binary.BigEndian.PutUint16(packet[1:3], uint16(len(topicB)))

	if _, err := conn.Write(packet); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to subscribe: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Subscribed to topic \"%s\" on broker %s (Ctrl+C to quit)\n", topic, *brokerAddr)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		fmt.Println("\nUnsubscribed. Goodbye!")
		os.Exit(0)
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Printf("→ %s\n", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Connection closed: %v\n", err)
	}
}
