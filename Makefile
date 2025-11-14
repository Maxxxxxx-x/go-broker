#!make

all:
	go build -o broker-server ./cmd/server
	go build -o broker-client ./cmd/client
