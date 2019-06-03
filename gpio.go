package gpio

import (
	"syscall"
	"unsafe"

	"github.com/juju/errors"
)

type Chip struct {
	fa              fdArc
	info            ChipInfo
	defaultConsumer string
}

func Open(path, defaultConsumer string) (*Chip, error) {
	fd, err := syscall.Open(path, syscall.O_RDWR|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	chip := &Chip{
		fa:              newFdArc(fd),
		defaultConsumer: defaultConsumer,
	}
	// runtime.SetFinalizer(chip, func(c *Chip) { c.Close() })
	err = RawGetChipInfo(chip.fa.fd, &chip.info)
	return chip, err
}

func (c *Chip) Close() error {
	c.fa.decref()
	return c.fa.wait()
}

func (c *Chip) Info() ChipInfo { return c.info }

func (c *Chip) OpenLines(flag RequestFlag, consumerLabel string, lines ...uint32) (*LinesHandle, error) {
	req := HandleRequest{
		Flags: flag,
		Lines: uint32(len(lines)),
	}
	copy(req.ConsumerLabel[:], []byte(consumerLabel))
	copy(req.LineOffsets[:], lines)

	c.fa.incref() // FIXME handle chip closed race
	err := RawGetLineHandle(c.fa.fd, &req)
	if err != nil {
		c.fa.decref()
		err = errors.Annotate(err, "GET_LINEHANDLE")
		return nil, err
	}
	if req.Fd <= 0 {
		c.fa.decref()
		err = errors.Errorf("GET_LINEHANDLE ioctl=success fd=%d", req.Fd)
		return nil, err
	}

	lh := &LinesHandle{
		chip:  c,
		fd:    req.Fd,
		count: req.Lines,
	}
	copy(lh.lines[:], req.LineOffsets[:])
	// runtime.SetFinalizer(lh, func(l *LinesHandle) { l.Close() })
	return lh, nil
}

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

type LinesHandle struct {
	chip   *Chip
	fd     int
	lines  [GPIOHANDLES_MAX]uint32
	values [GPIOHANDLES_MAX]byte
	count  uint32
}

func (self *LinesHandle) Close() error {
	err := syscall.Close(self.fd)
	self.chip.fa.decref()
	return err
}

func (self *LinesHandle) LineOffsets() []uint32 { return self.lines[:self.count] }

func (self *LinesHandle) Read() (HandleData, error) {
	data := HandleData{}
	err := RawGetLineValues(self.fd, &data)
	return data, err
}

func (self *LinesHandle) Write(bs ...byte) error {
	copy(self.values[:], bs)
	data := HandleData{Values: self.values}
	return RawSetLineValues(self.fd, &data)
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
