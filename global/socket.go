package global

import (
	"fmt"
	"syscall"
)

var CTSocket = new(ctSocket)

type ctSocket struct {
}

// connSocketEventHandler 调用读/写方法
func (c *ctSocket) aeHandler(el *AeEventLoop, fd int, clientData interface{}, mask int) {
	conn := clientData.(*connection)
	if mask&AeReadable != 0 && conn.readHandler != nil {
		conn.readHandler(conn)
	}
	if mask&AeWriteAble != 0 && conn.writeHandler != nil {
		conn.writeHandler(conn)
	}
}

// connSocketWrite
func (c *ctSocket) write(conn *connection, data []byte) {
	_, _ = syscall.Write(conn.fd, data)
}

// connSocketRead
func (c *ctSocket) read(conn *connection) []byte {
	buf := make([]byte, 500)
	n, err := syscall.Read(conn.fd, buf)
	fmt.Printf("read n: %v|err: %v\n", n, err)
	return buf
}

// connSocketClose
func (c *ctSocket) close(conn *connection) {
	if conn.fd != -1 {
		Server.el.AeDeleteFileEvent(conn.fd, AeReadable|AeWriteAble)
		syscall.Close(conn.fd)
		conn.fd = -1
	}
}

// connSocketListen
func (c *ctSocket) listen(listener *connListener) int {
	err := listener.listenToPort()
	if err != nil {
		return CErr
	}
	return COk
}

func (c *ctSocket) setReadHandler(conn *connection, cFunc ConnectionCallbackFunc) int {
	if conn.readHandler != nil {
		return COk
	}
	conn.readHandler = cFunc
	return Server.el.AeCreateFileEvent(conn.fd, AeReadable, conn.connType.aeHandler, conn)
}

// connSocketAcceptHandler
func (c *ctSocket) acceptHandler(el *AeEventLoop, fd int, privData interface{}, mask int) {
	//max := 10
	//for max > 0 {
	//	max--
	//	cfd, err := anetTcpAccept(fd, Server.bindAddress, Server.port)
	//	if err != nil {
	//		return
	//	}
	//	fmt.Printf("accept cfd: %v\n", cfd)
	//	acceptCommonHandler(connCreateAcceptedSocket(cfd), 0, Server.bindAddress)
	//}
	cfd, err := anetTcpAccept(fd, Server.bindAddress, Server.port)
	if err != nil {
		return
	}
	fmt.Printf("accept cfd: %v\n", cfd)
	acceptCommonHandler(connCreateAcceptedSocket(cfd), 0, Server.bindAddress)
}

func connCreateAcceptedSocket(fd int) *connection {
	conn := connCreateSocket()
	conn.fd = fd
	return conn
}

func connCreateSocket() *connection {
	conn := new(connection)
	conn.connType = CTSocket
	conn.fd = -1
	return conn
}
