package global

import (
	"fmt"
	"syscall"
)

const (
	COk  = 0
	CErr = -1
)

var Server = new(RedisServer)

type RedisServer struct {
	pid                     int
	port                    int
	bindAddress             string
	el                      *AeEventLoop
	statTotalReadsProcessed int
	tcpKeepAlive            int
	listener                *connListener
	currentClient           *client
}

type client struct {
	conn     *connection
	queryBuf string
}

type redisCommand struct {
}

func lookupCommand() {

}

func (s *RedisServer) initListener() int {
	listener := Server.listener
	listener.bindAddr = s.bindAddress
	listener.port = s.port
	listener.ct = CTSocket
	if listener.ct.listen(listener) == CErr {
		fmt.Printf("listen err\n")
		return CErr
	}
	fmt.Printf("listener fd: %v\n", listener.fd)
	return s.createSocketAcceptHandler(listener, listener.ct.acceptHandler)
}

func (s *RedisServer) createSocketAcceptHandler(listener *connListener, acceptHandler AeFileProc) int {
	err := s.el.AeCreateFileEvent(listener.fd, AeReadable, acceptHandler, listener)
	if err == AeErr {
		return CErr
	}
	return COk
}

func (s *RedisServer) initServer() int {
	Server.pid = syscall.Getpid()
	Server.port = 6379
	Server.bindAddress = "127.0.0.1"
	Server.el = AeCreateEventLoop(10)
	Server.tcpKeepAlive = 60
	Server.listener = new(connListener)
	return 0
}

func ServerRun() int {
	err := Server.initServer()
	if err != COk {
		return err
	}
	fmt.Printf("server init\n")
	err = Server.initListener()
	if err != COk {
		return err
	}
	fmt.Printf("listener init\n")
	Server.el.AeMain()
	return 0
}
