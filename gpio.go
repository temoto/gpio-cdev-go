package gpio

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"

	"github.com/juju/errors"
)

type Chip struct {
	f    *os.File
	fd   uintptr
	info ChipInfo
}

func Open(path string) (*Chip, error) {
	f, err := os.OpenFile(path, os.O_RDWR|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	chip := &Chip{
		f:  f,
		fd: f.Fd(),
	}
	err = ioctl(chip.fd, GPIO_GET_CHIPINFO_IOCTL, uintptr(unsafe.Pointer(&chip.info)))
	return chip, err
}

func (c *Chip) Close() error { return c.f.Close() }

func (c *Chip) Info() ChipInfo { return c.info }

func (c *Chip) EventLoop(line uint32, flag HandleFlag, events EventFlag, consumerLabel string, out chan<- EventData, stop <-chan struct{}) error {
	req := EventRequest{
		LineOffset:  line,
		HandleFlags: flag,
		EventFlags:  events,
	}
	copy(req.ConsumerLabel[:], []byte(consumerLabel))
	err := ioctl(c.fd, GPIO_GET_LINEEVENT_IOCTL, uintptr(unsafe.Pointer(&req)))
	if err != nil {
		err = errors.Trace(err)
		return err
	}

	for {
		log.Printf("req.Fd=%d", req.Fd)
		e, err := readEvent(req.Fd)
		if err != nil {
			err = errors.Trace(err)
			return err
		}
		select {
		case out <- e:
		case <-stop:
			return nil
		}
	}
}

func ioctl(fd, op, arg uintptr) error {
	r, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, op, arg)
	if errno != 0 {
		err := os.NewSyscallError("SYS_IOCTL", errno)
		log.Printf("ioctl fd=%d op=%x arg=%x err=%v", fd, op, arg, err)
		return err
	} else if r != 0 {
		err := fmt.Errorf("SYS_IOCTL r=%d", r)
		log.Printf("ioctl fd=%d op=%x arg=%x err=%v", fd, op, arg, err)
		return err
	}
	return nil
}

func readEvent(fd uintptr) (EventData, error) {
	const esz = int(unsafe.Sizeof(EventData{}))
	type eventBuf [esz]byte
	var buf [esz]byte
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
