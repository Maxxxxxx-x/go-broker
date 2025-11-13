package reactor

import (
	"log"
	"syscall"

	"github.com/Maxxxxxx-x/go-broker/internal/transport"
	"github.com/Maxxxxxx-x/go-broker/internal/types"
)

const MAX_EVENTS = 32

type Reactor struct {
	epFd      int
	listenFd  int
	eventFd   int
	logicChan chan<- types.Event
	respChan  <-chan types.Response
	buffers   map[int][]byte
	pending   map[int][]byte
}

func NewReactor(epFd, listenFd, eventFd int, logicChan chan types.Event, respChan chan types.Response) *Reactor {
	return &Reactor{
		epFd:      epFd,
		listenFd:  listenFd,
		eventFd:   eventFd,
		logicChan: logicChan,
		respChan:  respChan,
		buffers:   make(map[int][]byte),
		pending:   make(map[int][]byte),
	}
}

func (reactor *Reactor) Run() {
	for {
		var events [MAX_EVENTS]syscall.EpollEvent
		n, err := syscall.EpollWait(reactor.epFd, events[:], -1)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			log.Printf("epoll_wait error: %v\n", err)
			continue
		}

		for i := range n {
			fd := int(events[i].Fd)
			event := events[i].Events

			if fd == reactor.listenFd {
				reactor.accept()
				continue
			}

			if fd == reactor.eventFd {
				transport.DrainEventFd(reactor.eventFd)
				reactor.drainResponse()
				continue
			}

			if event&syscall.EPOLLOUT != 0 {
				reactor.write(fd)
				if _, ok := reactor.buffers[fd]; !ok {
					continue
				}
			}

			if event&(syscall.EPOLLIN|syscall.EPOLLHUP) != 0 {
				reactor.read(fd)
			}
		}
	}
}
