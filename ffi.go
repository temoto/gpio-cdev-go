package gpio

import (
	"fmt"
)

type ChipInfo struct {
	Name  [32]byte
	Label [32]byte
	Lines uint32
}

func (self *ChipInfo) String() string {
	return fmt.Sprintf("name=%s label=%s lines=%d",
		cstr(self.Name[:]), cstr(self.Label[:]), self.Lines)
}

type LineFlag uint32

const (
	GPIOLINE_FLAG_KERNEL      LineFlag = 1 << 0
	GPIOLINE_FLAG_IS_OUT      LineFlag = 1 << 1
	GPIOLINE_FLAG_ACTIVE_LOW  LineFlag = 1 << 2
	GPIOLINE_FLAG_OPEN_DRAIN  LineFlag = 1 << 3
	GPIOLINE_FLAG_OPEN_SOURCE LineFlag = 1 << 4
)

type LineInfo struct {
	LineOffset uint32
	Flags      uint32
	Name       [32]byte
	Consumer   [32]byte
}

const GPIOHANDLES_MAX = 64

type HandleFlag uint32

const (
	GPIOHANDLE_REQUEST_INPUT       HandleFlag = 1 << 0
	GPIOHANDLE_REQUEST_OUTPUT      HandleFlag = 1 << 1
	GPIOHANDLE_REQUEST_ACTIVE_LOW  HandleFlag = 1 << 2
	GPIOHANDLE_REQUEST_OPEN_DRAIN  HandleFlag = 1 << 3
	GPIOHANDLE_REQUEST_OPEN_SOURCE HandleFlag = 1 << 4
)

// Information about a GPIO handle request
type HandleRequest struct {
	// an array of desired lines, specified by offset index for the associated GPIO device
	LineOffsets [GPIOHANDLES_MAX]uint32

	// desired flags for the desired GPIO lines, such as
	// GPIOHANDLE_REQUEST_OUTPUT, GPIOHANDLE_REQUEST_ACTIVE_LOW etc, OR:ed
	// together. Note that even if multiple lines are requested, the same flags
	// must be applicable to all of them, if you want lines with individual
	// flags set, request them one by one. It is possible to select
	// a batch of input or output lines, but they must all have the same
	// characteristics, i.e. all inputs or all outputs, all active low etc
	Flags HandleFlag

	// if the GPIOHANDLE_REQUEST_OUTPUT is set for a requested
	// line, this specifies the default output value, should be 0 (low) or
	// 1 (high), anything else than 0 or 1 will be interpreted as 1 (high)
	// @consumer_label: a desired consumer label for the selected GPIO line(s)
	// such as "my-bitbanged-relay"
	DefaultValues [GPIOHANDLES_MAX]byte

	// a desired consumer label for the selected GPIO line(s)
	// such as "my-bitbanged-relay"
	ConsumerLabel [32]byte

	// number of lines requested in this request, i.e. the number of
	// valid fields in the above arrays, set to 1 to request a single line
	Lines uint32

	// if successful this field will contain a valid anonymous file handle
	// after a GPIO_GET_LINEHANDLE_IOCTL operation, zero or negative value
	// means error
	Fd uintptr
}

type HandleData struct {
	// when getting the state of lines this contains the current
	// state of a line, when setting the state of lines these should contain
	// the desired target state
	Values [GPIOHANDLES_MAX]byte
}

type EventFlag uint32

const (
	GPIOEVENT_REQUEST_RISING_EDGE  EventFlag = (1 << 0)
	GPIOEVENT_REQUEST_FALLING_EDGE EventFlag = (1 << 1)
	GPIOEVENT_REQUEST_BOTH_EDGES   EventFlag = ((1 << 0) | (1 << 1))
)

type EventRequest struct {
	// the desired line to subscribe to events from, specified by
	// offset index for the associated GPIO device
	LineOffset uint32

	// desired handle flags for the desired GPIO line, such as
	// GPIOHANDLE_REQUEST_ACTIVE_LOW or GPIOHANDLE_REQUEST_OPEN_DRAIN
	HandleFlags HandleFlag

	// desired flags for the desired GPIO event line, such as
	// GPIOEVENT_REQUEST_RISING_EDGE or GPIOEVENT_REQUEST_FALLING_EDGE
	EventFlags EventFlag

	// a desired consumer label for the selected GPIO line(s)
	// such as "my-listener"
	ConsumerLabel [32]byte

	// if successful this field will contain a valid anonymous file handle
	// after a GPIO_GET_LINEEVENT_IOCTL operation, zero or negative value means error
	Fd uintptr
}

type EventID uint32

const (
	GPIOEVENT_EVENT_RISING_EDGE  = 0x01
	GPIOEVENT_EVENT_FALLING_EDGE = 0x02
)

type EventData struct {
	Timestamp uint64
	ID        EventID
	_pad      uint32 //lint:ignore U1000 .
}

func cstr(bs []byte) string {
	length := 0
	for _, b := range bs {
		if b == 0 {
			break
		}
		length++
	}
	return string(bs[:length])
}
