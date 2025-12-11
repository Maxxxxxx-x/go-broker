#!make

all: clean
	go build -o ./bin/broker ./cmd/broker
	go build -o ./bin/publisher ./cmd/publisher
	go build -o ./bin/subscriber ./cmd/subscriber

clean:
	rm -rf ./bin
