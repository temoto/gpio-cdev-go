package gpio

import "unsafe"

// From <include/uapi/linux/gpio.h
var (
	GPIO_GET_CHIPINFO_IOCTL  uintptr = IoR(0xb4,0x01,unsafe.Sizeof(ChipInfo{}))
	GPIO_GET_LINEINFO_IOCTL          = IoWR(0xb4,0x02,unsafe.Sizeof(LineInfo{}))
	GPIO_GET_LINEHANDLE_IOCTL        = IoWR(0xb4,0x03,unsafe.Sizeof(HandleRequest{}))
	GPIO_GET_LINEEVENT_IOCTL         = IoWR(0xb4,0x04,unsafe.Sizeof(EventRequest{}))
	GPIOHANDLE_GET_LINE_VALUES_IOCTL = IoWR(0xb4,0x08,unsafe.Sizeof(HandleData{}))
	GPIOHANDLE_SET_LINE_VALUES_IOCTL = IoWR(0xb4,0x09,unsafe.Sizeof(HandleData{}))
)
