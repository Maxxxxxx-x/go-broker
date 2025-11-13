package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: client <host:port>")
		os.Exit(1)
	}
	addr := os.Args[1]

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("Connection failed: %v\n", err)
	}
	defer conn.Close()

	fmt.Println("Connected! Send \"goodbye\" to quit.")

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Printf("Server: %s\n", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Read error: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if _, err := conn.Write([]byte(line + "\n")); err != nil {
			log.Printf("Write error: %v\n", err)
			break
		}
		if strings.TrimSpace(line) == "goodbye" {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Input error: %v\n", err)
	}
}
