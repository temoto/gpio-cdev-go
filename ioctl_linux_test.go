// +build linux

package gpio

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"unsafe"
)

func TestIoctlHelper(t *testing.T) {
	cases := []struct {
		name   string
		expect uintptr
		actual uintptr
	}{
		{"GPIO_GET_CHIPINFO_IOCTL", GPIO_GET_CHIPINFO_IOCTL, ioR(0xb4, 0x01, unsafe.Sizeof(ChipInfo{}))},
		{"GPIO_GET_LINEINFO_IOCTL", GPIO_GET_LINEINFO_IOCTL, ioWR(0xb4, 0x02, unsafe.Sizeof(LineInfo{}))},
		{"GPIO_GET_LINEHANDLE_IOCTL", GPIO_GET_LINEHANDLE_IOCTL, ioWR(0xb4, 0x03, unsafe.Sizeof(HandleRequest{}))},
		{"GPIO_GET_LINEEVENT_IOCTL", GPIO_GET_LINEEVENT_IOCTL, ioWR(0xb4, 0x04, unsafe.Sizeof(EventRequest{}))},
		{"GPIOHANDLE_GET_LINE_VALUES_IOCTL", GPIOHANDLE_GET_LINE_VALUES_IOCTL, ioWR(0xb4, 0x08, unsafe.Sizeof(HandleData{}))},
		{"GPIOHANDLE_SET_LINE_VALUES_IOCTL", GPIOHANDLE_SET_LINE_VALUES_IOCTL, ioWR(0xb4, 0x09, unsafe.Sizeof(HandleData{}))},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expect, c.actual, c.name)
		})
	}
}
