package global

import (
	"fmt"
	"time"
)

const (
	MaxAcceptsPerCall = 1000
)

func acceptCommonHandler(conn *connection, flags int, ip string) {
	c := createClient(conn)
	if c == nil {
		return
	}

}

func createClient(conn *connection) *client {
	c := new(client)
	if conn != nil {
		conn.connEnableTcpNoDelay()
		conn.connKeepAlive(Server.tcpKeepAlive)
		conn.connectSetReadHandler(readQueryFromClient)
		conn.connSetPrivateData(c)
	}
	c.conn = conn
	fmt.Printf("client created\n")
	return c
}

func readQueryFromClient(conn *connection) {
	_client := conn.connGetPrivateData().(*client)
	// TODO 推迟
	Server.statTotalReadsProcessed++
	fmt.Printf("%s read from client|current conn fd: %v\n", time.Now().Format("2006-01-02 15:04:06"), conn.fd)
	buf := conn.connType.read(conn)
	_client.queryBuf = string(buf)
	// TODO 处理事件
	fmt.Printf("read from client: %v\n", string(buf))
}

func processInputBuffer(c *client) int {
	//TODO
	return 0
}

func processCommand(c *client) int {
	// TODO
	return 0
}
