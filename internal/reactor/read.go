package reactor

import (
	"log"
	"syscall"

	"github.com/Maxxxxxx-x/go-broker/internal/transport"
	"github.com/Maxxxxxx-x/go-broker/internal/types"
)

func (reactor *Reactor) read(fd int) {
	tmp := make([]byte, 1024)

	for {
		n, err := syscall.Read(fd, tmp)
		if n > 0 {
			reactor.buffers[fd] = append(reactor.buffers[fd], tmp[:n]...)

			for {
				buf := reactor.buffers[fd]
				line, ok := transport.ExtractLine(&buf)
				if !ok {
					break
				}
				reactor.buffers[fd] = buf

				select {
				case reactor.logicChan <- types.Event{Fd: fd, Message: line}:
				default:
					log.Printf("logic channel full, dropping message from fd=%d\n", fd)
				}
			}
			return
		}

		if n == 0 {
			if _, hasPending := reactor.pending[fd]; !hasPending {
				reactor.close(fd)
			}
			return
		}

		if err != syscall.EAGAIN && err != syscall.EINTR {
			log.Printf("read error on fd=%d: %v\n", fd, err)
			reactor.close(fd)
			return
		}
	}
}
