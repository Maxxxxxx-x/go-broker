package transport

import "syscall"

const (
	EPOLL_IN  = syscall.EPOLLIN
	EPOLL_OUT = syscall.EPOLLOUT
	EPOLL_ET  = 1 << 30
)

func NewEpoll() (int, error) {
	return syscall.EpollCreate1(0)
}

func EpollCtlAdd(epfd, fd int, events uint32) error {
	var epollEvent syscall.EpollEvent
	epollEvent.Events = events
	epollEvent.Fd = int32(fd)
	return syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &epollEvent)
}

func EpollCtlMod(epfd, fd int, events uint32) error {
	var epollEvent syscall.EpollEvent
	epollEvent.Events = events
	epollEvent.Fd = int32(fd)
	return syscall.EpollCtl(epfd, syscall.EPOLL_CTL_MOD, fd, &epollEvent)
}

func EpollCtlDel(epfd, fd int) error {
	return syscall.EpollCtl(epfd, syscall.EPOLL_CTL_DEL, fd, nil)
}
