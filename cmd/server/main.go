package main

import (
	"flag"
	"log"

	"github.com/Maxxxxxx-x/go-broker/internal/server"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "server port")
	flag.Parse()

	if port < 1 || port > 65535 {
		log.Fatalf("Invalid port: %d (must be 1-65535)\n", port)
	}

	log.Printf("Starting reactor server on :%d\n", port)
	if err := server.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}
