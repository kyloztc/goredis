package global

import (
	"fmt"
	"testing"
)

func TestAeCreateEventLoop(t *testing.T) {
	eventLoop := AeCreateEventLoop(10)
	fe := &eventLoop.events[1]
	fe.mask = 2
	fmt.Printf("%v\n", eventLoop.events[1].mask)
}

func testProc(eventLoop *AeEventLoop, fd int, clientData interface{}, mask int) {
	return
}
