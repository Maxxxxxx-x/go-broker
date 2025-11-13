package reactor

import (
	"log"
	"syscall"

	"github.com/Maxxxxxx-x/go-broker/internal/transport"
)

func (reactor *Reactor) accept() {
	for {
		nfd, _, err := syscall.Accept(reactor.listenFd)
		if err != nil {
			if err == syscall.EAGAIN || err == syscall.EINTR {
				return
			}
			log.Printf("accept error: %v\n", err)
			continue
		}

		if err := syscall.SetNonblock(nfd, true); err != nil {
			log.Printf("Setnonblock error on fd=%d: %v", nfd, err)
			syscall.Close(nfd)
			continue
		}

		if err := transport.EpollCtlAdd(reactor.epFd, nfd, transport.EPOLL_IN|transport.EPOLL_ET); err != nil {
			log.Printf("epoll add error on fd=%d: %v", nfd, err)
			syscall.Close(nfd)
			continue
		}

		reactor.buffers[nfd] = nil
		log.Printf("client connected: fd=%d\n", nfd)
	}
}
