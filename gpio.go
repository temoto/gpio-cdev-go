package gpio

import (
	"fmt"
	"syscall"

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

type LineSetFunc func(value byte)

// offset -> idx in self.lines/values
func (self *LinesHandle) mustFindLine(line uint32) int {
	for i, l := range self.lines {
		if uint32(i) >= self.count {
			break
		}
		if l == line {
			return i
		}
	}
	panic(fmt.Sprintf("code error invalid line=%d registered=%v", line, self.LineOffsets()))
}

func (self *LinesHandle) SetFunc(line uint32) LineSetFunc {
	idx := self.mustFindLine(line)
	return func(value byte) {
		self.values[idx] = value
	}
}

func (self *LinesHandle) LineOffsets() []uint32 { return self.lines[:self.count] }

func (self *LinesHandle) Read() (HandleData, error) {
	data := HandleData{}
	err := RawGetLineValues(self.fd, &data)
	return data, err
}

func (self *LinesHandle) Flush() error {
	data := HandleData{Values: self.values}
	return RawSetLineValues(self.fd, &data)
}

func (self *LinesHandle) SetBulk(bs ...byte) { copy(self.values[:], bs) }
