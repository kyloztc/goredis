package main

import (
	"fmt"
	"goredis/global"
	"net"
	"syscall"
)

type Socket struct {
	FileDescriptor int
}

func Listen(ip string, port int) (*Socket, error) {
	socket := &Socket{}

	socketFileDescriptor, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket (%v)", err)
	}

	fmt.Printf("fd: %v\n", socketFileDescriptor)

	socket.FileDescriptor = socketFileDescriptor

	socketAddress := &syscall.SockaddrInet4{Port: port}
	copy(socketAddress.Addr[:], net.ParseIP(ip))

	if err = syscall.Bind(socket.FileDescriptor, socketAddress); err != nil {
		return nil, fmt.Errorf("failed to bind socket (%v)", err)
	}

	if err = syscall.Listen(socket.FileDescriptor, syscall.SOMAXCONN); err != nil {
		return nil, fmt.Errorf("failed to listen on socket (%v)", err)
	}

	return socket, nil
}

func main() {
	global.ServerRun()
}
