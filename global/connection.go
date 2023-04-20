package global

type ConnectionState int

const (
	ConnStateNone = iota
	ConnStateConnecting
	ConnStateAccepting
	ConnStateConnected
	ConnStateClosed
	ConnStateError
)

type ConnectionCallbackFunc func(conn *connection)

type ConnectionType interface {
	acceptHandler(el *AeEventLoop, fd int, privData interface{}, mask int)
	listen(listener *connListener) int
	aeHandler(el *AeEventLoop, fd int, clientData interface{}, mask int)
	write(conn *connection, data []byte)
	read(conn *connection) []byte
	close(conn *connection)
	setReadHandler(conn *connection, cFunc ConnectionCallbackFunc) int
}

type connection struct {
	connType     ConnectionType
	state        ConnectionState
	flags        int
	refs         int
	privateData  interface{}
	connHandler  ConnectionCallbackFunc
	writeHandler ConnectionCallbackFunc
	readHandler  ConnectionCallbackFunc
	fd           int
}

type connListener struct {
	fd       int
	bindAddr string
	port     int
	ct       ConnectionType
}

func connCreateAcceptSocket(fd int) *connection {
	conn := connCreateSocket()
	conn.fd = fd
	conn.state = ConnStateAccepting
	return conn
}

func (c *connection) connGetState() ConnectionState {
	return c.state
}

func (c *connection) connClose() {
	c.connType.close(c)
}

func (c *connection) connEnableTcpNoDelay() int {
	if c.fd == -1 {
		return CErr
	}
	anetEnableTcpNoDelay(c.fd)
	return COk
}

func (c *connection) connKeepAlive(interval int) int {
	if c.fd == -1 {
		return CErr
	}
	anetKeepAlive(c.fd, interval)
	return COk
}

func (c *connection) connGetPrivateData() interface{} {
	return c.privateData
}

func (l *connListener) listenToPort() error {
	fd, err := anetTcpServer(l.port, l.bindAddr)
	if err != nil {
		return err
	}
	l.fd = fd

	return nil
}

func (c *connection) connectSetReadHandler(cFunc ConnectionCallbackFunc) int {
	return c.connType.setReadHandler(c, cFunc)
}

func (c *connection) connSetPrivateData(data interface{}) {
	c.privateData = data
}

func (c *connection) connAccept(cFunc ConnectionCallbackFunc) {

}
