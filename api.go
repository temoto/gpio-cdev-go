package gpio

import (
	"errors"
	"io"
	"time"
)

var ErrClosed = errors.New("already closed")

// Please use this to check whether gpio.*.Close() was already called.
func IsClosed(err error) bool {
	return err == ErrClosed
}

type Chiper interface {
	io.Closer
	Info() ChipInfo
	OpenLines(flag RequestFlag, consumerLabel string, lines ...uint32) (Lineser, error)
	GetLineEvent(line uint32, flag RequestFlag, events EventFlag, consumerLabel string) (Eventer, error)
}

type LineSetFunc func(value byte)

type Lineser interface {
	io.Closer
	SetFunc(line uint32) LineSetFunc
	LineOffsets() []uint32
	Read() (HandleData, error)
	Flush() error
	SetBulk(bs ...byte)
}

type Eventer interface {
	io.Closer
	Read() (byte, error)
	Wait(timeout time.Duration) (EventData, error)
}

// compile-time interface check
var _ Chiper = &chip{}
