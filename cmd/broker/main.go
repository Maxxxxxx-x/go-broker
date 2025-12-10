package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"

	broker "github.com/Maxxxxxx-x/go-broker/internal"
)

func main() {
	addr := flag.String("addr", ":8080", "TCP Address to listen to")
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen on %s: %v\n", *addr, err)
		os.Exit(1)
	}
	defer listener.Close()

	b := broker.New()
	fmt.Printf("Broker listening on %s\n", listener.Addr())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		fmt.Printf("\nShutting down broker...\n")
		listener.Close()
		os.Exit(0)
	}()

	if err := b.Serve(listener); err != nil {
		fmt.Fprintf(os.Stderr, "Broker error: %v\n", err)
	}
}
