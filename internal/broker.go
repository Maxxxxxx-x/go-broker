package broker

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
)

const (
	CTRL_SUBSCRIBE byte = 1
	CTRL_PUBLISH   byte = 2
)

type Broker struct {
	subscribers map[string][]net.Conn
	mutx        sync.RWMutex
}

func New() *Broker {
	return &Broker{
		subscribers: make(map[string][]net.Conn),
	}
}

func (broker *Broker) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go broker.handleConn(conn)
	}
}

func (broker *Broker) handleConn(conn net.Conn) {
	defer conn.Close()

	ctrlByte := make([]byte, 1)
	if _, err := io.ReadFull(conn, ctrlByte); err != nil {
		return
	}

	lenBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return
	}
	topicLen := binary.BigEndian.Uint16(lenBuf)

	topicBuf := make([]byte, topicLen)
	if _, err := io.ReadFull(conn, topicBuf); err != nil {
		return
	}
	topic := string(topicBuf)

	switch ctrlByte[0] {
	case CTRL_SUBSCRIBE:
		broker.addSubscriber(topic, conn)
		io.Copy(io.Discard, conn)

	case CTRL_PUBLISH:
		payloadLenBuf := make([]byte, 4)
		if _, err := io.ReadFull(conn, payloadLenBuf); err != nil {

			return
		}
		payloadLen := binary.BigEndian.Uint32(payloadLenBuf)

		payload := make([]byte, payloadLen)
		if _, err := io.ReadFull(conn, payload); err != nil {
			return
		}

		broker.broadcast(topic, payload)

	default:
		return
	}
}

func (broker *Broker) addSubscriber(topic string, conn net.Conn) {
	broker.mutx.Lock()
	broker.subscribers[topic] = append(broker.subscribers[topic], conn)
	broker.mutx.Unlock()

	go func() {
		<-closeConn(conn)
		broker.removeSubscriber(topic, conn)
	}()
}

func (broker *Broker) removeSubscriber(topic string, conn net.Conn) {
	broker.mutx.Lock()
	defer broker.mutx.Unlock()

	if list, ok := broker.subscribers[topic]; ok {
		for idx, c := range list {
			if c == conn {
				broker.subscribers[topic] = append(list[:idx], list[idx+1:]...)
				break
			}
		}
		if len(broker.subscribers[topic]) == 0 {
			delete(broker.subscribers, topic)
		}
	}
}

func (broker *Broker) broadcast(topic string, msg []byte) {
	broker.mutx.RLock()
	defer broker.mutx.RUnlock()

	if subs, ok := broker.subscribers[topic]; ok {
		for _, sub := range subs {
			go func(conn net.Conn) {
				_, _ = conn.Write(append(msg, '\n'))
			}(sub)
		}
	}
}

func closeConn(conn net.Conn) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		io.Copy(io.Discard, conn)
		close(ch)
	}()

	return ch
}
