// +build linux

package gpio

import (
	"testing"
	"unsafe"
	"github.com/stretchr/testify/assert"
)

const TEST_GPIO_GET_CHIPINFO_IOCTL=uintptr(0x8044b401)
const TEST_GPIO_GET_LINEINFO_IOCTL=uintptr(0xc048b402)
const TEST_GPIO_GET_LINEHANDLE_IOCTL=uintptr(0xc16cb403)
const TEST_GPIO_GET_LINEEVENT_IOCTL=uintptr(0xc030b404)
const TEST_GPIOHANDLE_GET_LINE_VALUES_IOCTL=uintptr(0xc040b408)
const TEST_GPIOHANDLE_SET_LINE_VALUES_IOCTL=uintptr(0xc040b409)


func TestIoctlHelper(t *testing.T) {
	a := assert.New(t)
	// #define GPIO_GET_CHIPINFO_IOCTL _IOR(0xB4, 0x01, struct gpiochip_info)
	a.Equal(
		TEST_GPIO_GET_CHIPINFO_IOCTL,
		IoR(0xb4,0x01,unsafe.Sizeof(ChipInfo{})),
		"compare GPIO_GET_CHIPINFO_IOCTL with predefined one",
	)
	// #define GPIO_GET_LINEINFO_IOCTL _IOWR(0xB4, 0x02, struct gpioline_info)
	a.Equal(
		TEST_GPIO_GET_LINEINFO_IOCTL,
		IoWR(0xb4,0x02,unsafe.Sizeof(LineInfo{})),
		"compare GPIO_GET_LINEINFO_IOCTL with predefined one",
	)
	// #define GPIO_GET_LINEHANDLE_IOCTL _IOWR(0xB4, 0x03, struct gpiohandle_request)
	a.Equal(
		TEST_GPIO_GET_LINEHANDLE_IOCTL,
		IoWR(0xb4,0x03,unsafe.Sizeof(HandleRequest{})),
		"compare GPIO_GET_LINEHANDLE_IOCTL with predefined one",
	)
	// #define GPIO_GET_LINEEVENT_IOCTL _IOWR(0xB4, 0x04, struct gpioevent_request)
	a.Equal(
		TEST_GPIO_GET_LINEEVENT_IOCTL,
		IoWR(0xb4,0x04,unsafe.Sizeof(EventRequest{})),
		"compare GPIO_GET_LINEEVENT_IOCTL with predefined one",
	)
	// #define GPIOHANDLE_GET_LINE_VALUES_IOCTL _IOWR(0xB4, 0x08, struct gpiohandle_data)
	a.Equal(
		TEST_GPIOHANDLE_GET_LINE_VALUES_IOCTL,
		IoWR(0xb4,0x08,unsafe.Sizeof(HandleData{})),
		"compare GPIOHANDLE_GET_LINE_VALUES_IOCTL with predefined one",
	)
	// #define GPIOHANDLE_SET_LINE_VALUES_IOCTL _IOWR(0xB4, 0x09, struct gpiohandle_data)
	a.Equal(
		TEST_GPIOHANDLE_SET_LINE_VALUES_IOCTL,
		IoWR(0xb4,0x09,unsafe.Sizeof(HandleData{})),
		"compare GPIOHANDLE_SET_LINE_VALUES_IOCTL with predefined one",
	)
}