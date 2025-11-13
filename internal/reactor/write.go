package reactor

import (
	"log"
	"syscall"

	"github.com/Maxxxxxx-x/go-broker/internal/transport"
)

func (reactor *Reactor) write(fd int) {
	data := reactor.pending[fd]
	n, err := syscall.Write(fd, data)
	if err != nil && err != syscall.EAGAIN && err != syscall.EINTR {
		reactor.close(fd)
		return
	}

	data = data[n:]
	if len(data) == 0 {
		delete(reactor.pending, fd)
		if err := transport.EpollCtlMod(reactor.epFd, fd, transport.EPOLL_IN|transport.EPOLL_ET); err != nil {
			log.Printf("epoll mod err on fd=%d: %v", fd, err)
		}
		return
	}

	reactor.pending[fd] = data
	if err := transport.EpollCtlMod(reactor.epFd, fd, transport.EPOLL_OUT|transport.EPOLL_ET); err != nil {
		log.Printf("epoll mod error on fd=%d: %v\n", fd ,err)
	}
}

func (reactor *Reactor) drainResponse() {
	for {
		select {
		case resp := <-reactor.respChan:
			if resp.Close {
				delete(reactor.pending, resp.Fd)
				reactor.close(resp.Fd)
				continue
			}
			data := []byte(resp.Message)
			if p, ok := reactor.pending[resp.Fd]; ok {
				data = append(p, data...)
				delete(reactor.pending, resp.Fd)
			}
			n, err := syscall.Write(resp.Fd, data)
			if err != nil && err != syscall.EAGAIN && err != syscall.EINTR {
				reactor.close(resp.Fd)
				continue
			}

			if n > 0 {
				log.Printf("responded to client fd=%d: %q (%d bytes)\n", resp.Fd, resp.Message[:len(resp.Message)-1], n)
			}

			if n < len(data) {
				reactor.pending[resp.Fd] = data[n:]
				if err := transport.EpollCtlMod(reactor.epFd, resp.Fd, transport.EPOLL_IN|transport.EPOLL_OUT|transport.EPOLL_ET); err != nil {
					log.Printf("epoll mod err on fd=%d: %v\n", resp.Fd, err)
				}
			}
		default:
			return
		}
	}
}
