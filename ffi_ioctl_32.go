package gpio

// Likely these constants are not portable, sorry, I don't know better.
const (
	GPIO_GET_CHIPINFO_IOCTL          = 0x8044b401
	GPIO_GET_LINEINFO_IOCTL          = 0xc048b402
	GPIO_GET_LINEHANDLE_IOCTL        = 0xc16cb403
	GPIO_GET_LINEEVENT_IOCTL         = 0xc030b404
	GPIOHANDLE_GET_LINE_VALUES_IOCTL = 0xc040b408
	GPIOHANDLE_SET_LINE_VALUES_IOCTL = 0xc040b409
)
