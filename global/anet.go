package global

import (
	"fmt"
	"net"
	"syscall"
)

func anetTcpServer(port int, bindAddr string) (int, error) {
	return _anetTcpServer(port, bindAddr, syscall.AF_INET)
}

func _anetTcpServer(port int, bindAddr string, af int) (int, error) {
	fd, err := syscall.Socket(af, syscall.SOCK_STREAM, 0)
	if err != nil {
		return -1, err
	}
	socketAddress := &syscall.SockaddrInet4{Port: port}
	copy(socketAddress.Addr[:], net.ParseIP(bindAddr))
	if err = syscall.Bind(fd, socketAddress); err != nil {
		return -1, fmt.Errorf("failed to bind socket (%v)", err)
	}

	if err = syscall.Listen(fd, syscall.SOMAXCONN); err != nil {
		return -1, fmt.Errorf("failed to listen on socket (%v)", err)
	}

	return fd, nil
}

func anetTcpAccept(s int, ip string, port int) (int, error) {
	fd, _, err := anetGenericAccept(s)
	return fd, err
}

func anetGenericAccept(s int) (int, *syscall.Sockaddr, error) {
	fd, sa, err := syscall.Accept(s)
	return fd, &sa, err
}

func anetEnableTcpNoDelay(fd int) error {
	return anetSetTcpNoDelay(fd, 1)
}

func anetSetTcpNoDelay(fd int, val int) error {
	err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, val)
	if err != nil {
		return err
	}
	return nil
}

func anetKeepAlive(fd int, interval int) error {
	val := 1
	err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, val)
	return err
}
