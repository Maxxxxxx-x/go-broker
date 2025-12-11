package broker

import (
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
)

const (
	CTRL_SUBSCRIBE byte = 1
	CTRL_PUBLISH   byte = 2
)

type Broker struct {
	subscribers map[string][]net.Conn
	mutx        sync.RWMutex
	connId      atomic.Uint64
}

func New() *Broker {
	broker := &Broker{
		subscribers: make(map[string][]net.Conn),
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(a.Value.Time().Format("2003/10/28 15:04:05")),
				}
			}

			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				label := map[slog.Level]string{
					slog.LevelDebug: "DEBUG",
					slog.LevelInfo:  "INFO ",
					slog.LevelWarn:  "WARN ",
					slog.LevelError: "ERROR",
				}[level]
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(label),
				}
			}

			return a
		},
	})))
	slog.Info("Broker initiated")
	return broker
}

func (broker *Broker) Serve(listener net.Listener) error {
	slog.Info("Broker starting", "listen_addr", listener.Addr())
	defer slog.Warn("Broker stopped")

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		connId := broker.connId.Add(1)
		slog.Info("New connection", "conn_id", connId, "remote", conn.RemoteAddr())

		go broker.handleConn(conn, connId)
	}
}

func (broker *Broker) handleConn(conn net.Conn, connId uint64) {
	defer func() {
		conn.Close()
		slog.Info("Connection closed", "conn_id", connId, "remote", conn.RemoteAddr())
	}()

	remote := conn.RemoteAddr().String()

	ctrlByte := make([]byte, 1)
	if _, err := io.ReadFull(conn, ctrlByte); err != nil {
		slog.Warn("Failed to read control byte", "conn_id", connId, "remote", remote, "error", err)
		return
	}

	lenBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		slog.Warn("Failed to read topic length", "conn_id", connId, "remote", remote, "error", err)
		return
	}
	topicLen := binary.BigEndian.Uint16(lenBuf)

	topicBuf := make([]byte, topicLen)
	if _, err := io.ReadFull(conn, topicBuf); err != nil {
		slog.Warn("Failed to read topic", "conn_id", connId, "remote", remote, "error", err)
		return
	}
	topic := string(topicBuf)

	switch ctrlByte[0] {
	case CTRL_SUBSCRIBE:
		slog.Info("SUBSCRIBE request", "conn_id", connId, "remote", remote, "topic", topic)
		broker.addSubscriber(topic, conn, connId, remote)
		io.Copy(io.Discard, conn)
		slog.Info("Subscriber disconnected (detected via io.Copy)", "conn_id", connId, "remopte", remote, "topic", topic)
		broker.removeSubscriber(topic, conn, connId, remote)

	case CTRL_PUBLISH:
		slog.Info("PUBLISH request", "conn_id", connId, "remote", remote, "topic", topic)

		payloadLenBuf := make([]byte, 4)
		if _, err := io.ReadFull(conn, payloadLenBuf); err != nil {
			slog.Warn("Failed to read payload length", "conn_id", connId, "error", err)
			return
		}
		payloadLen := binary.BigEndian.Uint32(payloadLenBuf)

		payload := make([]byte, payloadLen)
		if _, err := io.ReadFull(conn, payload); err != nil {
			slog.Warn("Failed to read payload", "conn_id", connId, "error", err)
			return
		}

		msg := string(payload)
		if len(msg) > 100 {
			msg = msg[:100] + "..."
		}
		slog.Info("Message published", "topic", topic, "size_bytes", payloadLen, "preview", msg)
		broker.broadcast(topic, payload, remote)

	default:
		slog.Warn("Unknown control packet", "conn_id", connId, "control", ctrlByte[0])
		return
	}
}

func (broker *Broker) addSubscriber(topic string, conn net.Conn, connId uint64, remote string) {
	broker.mutx.Lock()
	broker.subscribers[topic] = append(broker.subscribers[topic], conn)
	broker.mutx.Unlock()

	slog.Info("Subscriber added", "conn_id", connId, "remote", remote, "topic", topic, "total_subscribers", len(broker.subscribers[topic]))
}

func (broker *Broker) removeSubscriber(topic string, conn net.Conn, connId uint64, remote string) {
	broker.mutx.Lock()
	defer broker.mutx.Unlock()

	if list, ok := broker.subscribers[topic]; ok {
		removed := false
		for idx, c := range list {
			if c == conn {
				broker.subscribers[topic] = append(list[:idx], list[idx+1:]...)
				removed = true
				break
			}
		}

		newCount := len(broker.subscribers[topic])
		if removed {
			slog.Warn("Subscriber REMOVED", "conn_id", connId, "rmeote", remote, "topic", topic, "remaining", newCount)
		}

		if newCount == 0 {
			delete(broker.subscribers, topic)
			slog.Info("Topic cleaned up (no subscribers left)", "topic", topic)
		}
	}
}

func (broker *Broker) broadcast(topic string, msg []byte, publisher string) {
	broker.mutx.RLock()
	defer broker.mutx.RUnlock()

	if subs, ok := broker.subscribers[topic]; ok {
		slog.Info("Broadcasting message", "topic", topic, "recipients", len(subs), "from", publisher)
		for _, sub := range subs {
			go func(conn net.Conn) {
				if _, err := conn.Write(append(msg, '\n')); err != nil {
					slog.Warn("Failed to deliver message to subscriber", "remote", conn.RemoteAddr(), "error", err)
				}
			}(sub)
		}
	} else {
		slog.Debug("No subscribers for topic (message dropped)", "topic", topic)
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
