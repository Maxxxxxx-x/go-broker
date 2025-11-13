package server

import (
	"github.com/Maxxxxxx-x/go-broker/internal/logic"
	"github.com/Maxxxxxx-x/go-broker/internal/reactor"
	"github.com/Maxxxxxx-x/go-broker/internal/transport"
	"github.com/Maxxxxxx-x/go-broker/internal/types"
)

func Start(port int) error {
	eventFd, err := transport.NewEventFd()
	if err != nil {
		return err
	}
	defer eventFd.Close()

	listenFd, err := transport.ListenTCP(port)
	if err != nil {
		return err
	}
	defer transport.CloseFd(listenFd)

	epfd, err := transport.NewEpoll()
	if err != nil {
		return err
	}
	defer transport.CloseFd(epfd)

	if err := transport.EpollCtlAdd(epfd, listenFd, transport.EPOLL_IN); err != nil {
		return err
	}

	if err := transport.EpollCtlAdd(epfd, eventFd.Fd, transport.EPOLL_IN|transport.EPOLL_ET); err != nil {
		return err
	}

	logicChan := make(chan types.Event, 100)
	respChan := make(chan types.Response, 100)

	go logic.NewHandler(logicChan, respChan, eventFd).Run()
	reactor.NewReactor(epfd, listenFd, eventFd.Fd, logicChan, respChan).Run()

	return nil
}
