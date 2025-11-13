package reactor

import (
	"log"
	"syscall"

	"github.com/Maxxxxxx-x/go-broker/internal/transport"
)

func (reactor *Reactor) close(fd int) {
	if _, ok := reactor.buffers[fd]; !ok {
		return
	}

	if err := transport.EpollCtlDel(reactor.epFd, fd); err != nil {
		log.Printf("epoll del err on fd=%d: %v\n", fd, err)
	}

	if err := syscall.Close(fd); err != nil {
		log.Printf("close fd=%d: %v\n", fd, err)
	}

	delete(reactor.buffers, fd)
	delete(reactor.pending, fd)
	log.Printf("client disconnected: fd=%d\n", fd)
}
