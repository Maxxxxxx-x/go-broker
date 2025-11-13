package logic

import (
	"log"

	"github.com/Maxxxxxx-x/go-broker/internal/transport"
	"github.com/Maxxxxxx-x/go-broker/internal/types"
)

const EXIT_MESSAGE = "goodbye"

type Handler struct {
	in      <-chan types.Event
	out     chan<- types.Response
	eventFd *transport.EventFd
}

func NewHandler(in <-chan types.Event, out chan<- types.Response, eventFd *transport.EventFd) *Handler {
	return &Handler{
		in:      in,
		out:     out,
		eventFd: eventFd,
	}
}

func (handler *Handler) Run() {
	for event := range handler.in {
		msg := transport.TrimLine(event.Message)
		resp := types.Response{Fd: event.Fd}

		if msg == EXIT_MESSAGE {
			resp.Close = true
			log.Printf("Client fd=%d: disconnect request received", event.Fd)
		} else {
			resp.Message = msg + "\n"
		}

		select {
		case handler.out <- resp:
			if err := handler.eventFd.Wake(); err != nil {
				log.Printf("eventfd wake failed: %v\n", err)
			}
		default:
			log.Printf("response channel full, drpping response for fd=%d\n", event.Fd)
		}
	}
}
