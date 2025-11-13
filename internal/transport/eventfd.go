package transport

import (
	"log"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type EventFd struct {
	Fd int
}

func NewEventFd() (*EventFd, error) {
	fd, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	return &EventFd{Fd: fd}, err
}

func (event *EventFd) Close() {
	unix.Close(event.Fd)
}

func (event *EventFd) Wake() error {
	var one uint64 = 1
	_, err := syscall.Write(event.Fd, (*[8]byte)(unsafe.Pointer(&one))[:])
	return err
}

func DrainEventFd(fd int) {
	var b [8]byte
	_, err := syscall.Read(fd, b[:])
	if err != nil && err != syscall.EAGAIN && err != syscall.EINTR {
		log.Printf("drain eventfd: %v", err)
	}
}
