package global

import (
	"fmt"
	"syscall"
	"time"
)

const (
	AeOk  = 0
	AeErr = -1

	AeNone      = 0
	AeReadable  = 1
	AeWriteAble = 2
	AeBarrier   = 3

	AeFileEvents = 1 << 0
	AeTimeEvents = 1 << 1
)

type AeApiState interface {
	AeApiResize(setSize int) int
	AeApiAddEvent(fd int, mask int) int
	AeApiDelEvent(fd int, mask int)
	AeApiPoll(eventLoop *AeEventLoop, tvp *syscall.Timeval) int
}

type AeFileProc func(eventLoop *AeEventLoop, fd int, clientData interface{}, mask int)

type aeFileEvent struct {
	mask       int
	rfileProc  AeFileProc
	wfileProc  AeFileProc
	clientData interface{}
}

type aeFiredEvent struct {
	fd   int
	mask int
}

type AeEventLoop struct {
	maxfd   int
	setSize int
	events  []aeFileEvent
	fired   []aeFiredEvent
	apiData AeApiState
	stop    int
}

func EvSet(ke *syscall.Kevent_t, ident uint64, filter int16, flags uint16, fflags uint32, data int64, udata *byte) {
	ke.Ident = ident
	ke.Filter = filter
	ke.Flags = flags
	ke.Fflags = fflags
	ke.Data = data
	ke.Udata = udata
}

func AeCreateEventLoop(setSize int) *AeEventLoop {
	eventLoop := new(AeEventLoop)
	eventLoop.events = make([]aeFileEvent, setSize)
	eventLoop.fired = make([]aeFiredEvent, setSize)
	eventLoop.setSize = setSize
	eventLoop.maxfd = -1
	//for i := 0; i < setSize; i++ {
	//	eventLoop.events[i].mask = AeNone
	//}
	KqueueAeApiCreate(eventLoop)
	return eventLoop
}

func (e *AeEventLoop) AeGetSetSize() int {
	return e.setSize
}

func (e *AeEventLoop) AeStop() {
	e.stop = 1
}

func (e *AeEventLoop) AeCreateFileEvent(fd int, mask int, proc AeFileProc, clientData interface{}) int {
	if fd >= e.setSize {
		return AeErr
	}
	fe := &e.events[fd]
	if e.apiData.AeApiAddEvent(fd, mask) == -1 {
		fmt.Printf("create file event err\n")
		return AeErr
	}
	fe.mask |= mask
	if mask&AeReadable != 0 {
		fe.rfileProc = proc
	}
	if mask&AeWriteAble != 0 {
		fe.wfileProc = proc
	}
	fe.clientData = clientData
	if fd > e.maxfd {
		e.maxfd = fd
	}
	return AeOk
}

func (e *AeEventLoop) AeDeleteFileEvent(fd int, mask int) {
	if fd >= e.setSize {
		return
	}
	fe := &e.events[fd]
	if fe.mask == AeNone {
		return
	}
	e.apiData.AeApiDelEvent(fd, mask)
	// 更新maxfd
	fe.mask = AeNone
	if fd == e.maxfd && fe.mask == AeNone {
		j := 0
		for j = e.maxfd - 1; j >= 0; j-- {
			if e.events[j].mask != AeNone {
				break
			}
		}
		e.maxfd = j
	}
}

func (e *AeEventLoop) AeGetFileEvents(fd int) int {
	if fd >= e.setSize {
		return 0
	}
	fe := &e.events[fd]
	return fe.mask
}

func (e *AeEventLoop) AeProcessEvents(flags int) int {
	processed := 0
	numEvents := 0

	if flags&AeTimeEvents == 0 && flags&AeFileEvents == 0 {
		return 0
	}
	fmt.Printf("maxfd: %v\n", e.maxfd)
	if e.maxfd != -1 {
		numEvents = e.apiData.AeApiPoll(e, nil)
		fmt.Printf("num events: %v\n", numEvents)
		for j := 0; j < numEvents; j++ {
			fe := &e.events[e.fired[j].fd]
			mask := e.fired[j].mask
			fd := e.fired[j].fd
			fired := 0
			if mask&AeReadable != 0 {
				fe.rfileProc(e, fd, fe.clientData, mask)
				fired++
				fmt.Printf("process read proc\n")
			}
			if mask&AeWriteAble != 0 {
				fe.wfileProc(e, fd, fe.clientData, mask)
				fired++
				fmt.Printf("process write proc\n")
			}
			processed++
		}
	}
	fmt.Printf("process end\n")
	return processed
}

func (e *AeEventLoop) AeMain() {
	for e.stop != 1 {
		e.AeProcessEvents(AeFileEvents)
		time.Sleep(time.Second * 1)
	}
}
