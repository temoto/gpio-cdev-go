package gpio

import (
	"fmt"
	"os"
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
	err = ioctl(chip.fa.fd, GPIO_GET_CHIPINFO_IOCTL, uintptr(unsafe.Pointer(&chip.info)))
	return chip, err
}

func (c *Chip) Close() error {
	c.fa.decref()
	return c.fa.wait()
}

func (c *Chip) Info() ChipInfo { return c.info }

// func (c *Chip) LineRead(line uint32, flag RequestFlag) (byte, error) {
// 	req := HandleRequest{
// 		Flags: GPIOHANDLE_REQUEST_INPUT | flag,
// 		Lines: 1,
// 	}
// 	copy(req.ConsumerLabel[:], c.defaultConsumer)
// 	c.fa.incref()
// 	defer c.fa.decref()
// 	err := ioctl(c.fa.fd, GPIO_GET_LINEHANDLE_IOCTL, uintptr(unsafe.Pointer(&req)))
// }

func (c *Chip) GetLineEvent(line uint32, flag RequestFlag, events EventFlag, consumerLabel string) (*LineEventHandle, error) {
	req := EventRequest{
		LineOffset:   line,
		RequestFlags: GPIOHANDLE_REQUEST_INPUT | flag,
		EventFlags:   events,
	}
	copy(req.ConsumerLabel[:], []byte(consumerLabel))
	c.fa.incref()
	err := ioctl(c.fa.fd, GPIO_GET_LINEEVENT_IOCTL, uintptr(unsafe.Pointer(&req)))
	if err != nil {
		c.fa.decref()
		err = errors.Trace(err)
		return nil, err
	}

	le := &LineEventHandle{
		chip:    c,
		eventFd: req.Fd,
		line:    line,
		reqFlag: req.RequestFlags,
		events:  req.EventFlags,
	}
	return le, nil
}

type LineEventHandle struct {
	chip    *Chip
	eventFd int
	line    uint32
	reqFlag RequestFlag
	events  EventFlag
}

func (self *LineEventHandle) Close() error {
	err := syscall.Close(self.eventFd)
	self.chip.fa.decref()
	return err
}

func (self *LineEventHandle) Read() (byte, error) {
	d, err := readLines(self.eventFd)
	if err != nil {
		err = errors.Annotate(err, "event.Read")
	}
	return d.Values[0], err
	// return self.chip.LineRead(self.line, self.reqFlag)
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

func ioctl(fd int, op, arg uintptr) error {
	r, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), op, arg)
	if errno != 0 {
		err := os.NewSyscallError("SYS_IOCTL", errno)
		// log.Printf("ioctl fd=%d op=%x arg=%x err=%v", fd, op, arg, err)
		return err
	} else if r != 0 {
		err := fmt.Errorf("SYS_IOCTL r=%d", r)
		// log.Printf("ioctl fd=%d op=%x arg=%x err=%v", fd, op, arg, err)
		return err
	}
	return nil
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

func readLines(fd int) (HandleData, error) {
	var d HandleData
	err := ioctl(fd, GPIOHANDLE_GET_LINE_VALUES_IOCTL, uintptr(unsafe.Pointer(&d)))
	return d, err
}
