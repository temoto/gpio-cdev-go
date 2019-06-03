package gpio

import (
	"syscall"
	"unsafe"

	"github.com/juju/errors"
)

func (c *Chip) GetLineEvent(line uint32, flag RequestFlag, events EventFlag, consumerLabel string) (*LineEventHandle, error) {
	req := EventRequest{
		LineOffset:   line,
		RequestFlags: GPIOHANDLE_REQUEST_INPUT | flag,
		EventFlags:   events,
	}
	copy(req.ConsumerLabel[:], []byte(consumerLabel))

	c.fa.incref() // FIXME handle chip closed race
	err := RawGetLineEvent(c.fa.fd, &req)
	if err != nil {
		c.fa.decref()
		err = errors.Trace(err)
		return nil, err
	}

	le := &LineEventHandle{
		chip:    c,
		eventFd: req.Fd,
		reqFlag: req.RequestFlags,
		events:  req.EventFlags,
	}
	// runtime.SetFinalizer(le, func(l *LineEventHandle) { l.Close() })
	return le, nil
}

type LineEventHandle struct {
	chip    *Chip
	eventFd int
	reqFlag RequestFlag
	events  EventFlag
}

func (self *LineEventHandle) Close() error {
	err := syscall.Close(self.eventFd)
	self.chip.fa.decref()
	return err
}

func (self *LineEventHandle) Read() (byte, error) {
	var data HandleData
	err := RawGetLineValues(self.eventFd, &data)
	if err != nil {
		err = errors.Annotate(err, "event.Read")
	}
	return data.Values[0], err
}

func (self *LineEventHandle) Wait( /*FIXME timeout time.Duration*/ ) (EventData, error) {
	// syscall.Select()

	// log.Printf("req.Fd=%d", req.Fd)
	e, err := readEvent(self.eventFd)
	if err != nil {
		err = errors.Annotate(err, "event.Wait")
	}
	return e, err
}

func readEvent(fd int) (EventData, error) {
	// ugly dance around syscall.Read []byte instead of *void
	const esz = int(unsafe.Sizeof(EventData{}))
	type eventBuf [esz]byte
	var buf eventBuf
	var e EventData
	n, err := syscall.Read(int(fd), buf[:])
	if err != nil {
		err = errors.Annotate(err, "readEvent")
		return e, err
	}
	if n != esz {
		err = errors.Errorf("readEvent fail n=%d expected=%d", n, esz)
		return e, err
	}
	eb := (*eventBuf)(unsafe.Pointer(&e))
	copy((*eb)[:], buf[:])
	return e, nil
}
