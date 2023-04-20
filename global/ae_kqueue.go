package global

import "syscall"

type KqueueAeApiState struct {
	kqfd       int
	events     []syscall.Kevent_t
	eventsMask []int
}

func KqueueAeApiCreate(eventLoop *AeEventLoop) int {
	state := new(KqueueAeApiState)
	state.events = make([]syscall.Kevent_t, eventLoop.setSize)
	kqfd, err := syscall.Kqueue()
	if err != nil {
		return -1
	}
	state.kqfd = kqfd
	state.eventsMask = make([]int, eventLoop.setSize)
	eventLoop.apiData = state
	return 0
}

func (k *KqueueAeApiState) AeApiResize(setSize int) int {
	k.events = make([]syscall.Kevent_t, setSize)
	k.eventsMask = make([]int, setSize)
	return 0
}

func (k *KqueueAeApiState) AeApiAddEvent(fd int, mask int) int {
	ke := new(syscall.Kevent_t)
	if mask&AeReadable != 0 {
		EvSet(ke, uint64(fd), syscall.EVFILT_READ, syscall.EV_ADD, 0, 0, nil)
	}
	if mask&AeWriteAble != 0 {
		EvSet(ke, uint64(fd), syscall.EVFILT_WRITE, syscall.EV_ADD, 0, 0, nil)
	}
	if rsp, err := syscall.Kevent(k.kqfd, []syscall.Kevent_t{*ke}, nil, nil); err != nil || rsp == -1 {
		return -1
	}
	return 0
}

func (k *KqueueAeApiState) AeApiDelEvent(fd int, mask int) {
	ke := new(syscall.Kevent_t)
	if mask&AeReadable != 0 {
		EvSet(ke, uint64(fd), syscall.EVFILT_READ, syscall.EV_DELETE, 0, 0, nil)
	}
	if mask&AeWriteAble != 0 {
		EvSet(ke, uint64(fd), syscall.EVFILT_WRITE, syscall.EV_DELETE, 0, 0, nil)
	}
	_, _ = syscall.Kevent(k.kqfd, []syscall.Kevent_t{*ke}, nil, nil)
}

func (k *KqueueAeApiState) AeApiPoll(eventLoop *AeEventLoop, tvp *syscall.Timeval) int {
	retval := 0
	numEvents := 0
	if tvp != nil {
		timeout := new(syscall.Timespec)
		timeout.Sec = tvp.Sec
		timeout.Nsec = int64(tvp.Usec * 1000)
		retval, _ = syscall.Kevent(k.kqfd, nil, k.events, timeout)
	} else {
		retval, _ = syscall.Kevent(k.kqfd, nil, k.events, nil)
	}
	if retval > 0 {
		for j := 0; j < retval; j++ {
			e := k.events[j]
			fd := e.Ident
			mask := 0
			if e.Filter == syscall.EVFILT_READ {
				mask = AeReadable
			} else if e.Filter == syscall.EVFILT_WRITE {
				mask = AeWriteAble
			}
			if mask != 0 {
				eventLoop.fired[numEvents].fd = int(fd)
				eventLoop.fired[numEvents].mask = mask
				numEvents++
			}
		}
	}
	return numEvents
}
