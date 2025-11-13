package transport

import "syscall"

func ListenTCP(port int) (int, error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return 0, err
	}

	if err = syscall.SetNonblock(fd, true); err != nil {
		return 0, err
	}

	addr := syscall.SockaddrInet4{Port: port}
	copy(addr.Addr[:], []byte{0, 0, 0, 0})

	if err = syscall.Bind(fd, &addr); err != nil {
		syscall.Close(fd)
		return 0, err
	}

	if err = syscall.Listen(fd, 10); err != nil {
		syscall.Close(fd)
		return 0, err
	}

	return fd, nil
}

func CloseFd(fd int) {
	syscall.Close(fd)
}
