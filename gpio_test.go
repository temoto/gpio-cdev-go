package gpio

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// validate and extract test setup
func getEnv(t *testing.T) (gpioDev string, gpioIn uint32, gpioOut uint32, hasGpio bool, hasLoopback bool) {
	gpioDev = os.Getenv("GPIO_TEST_DEV")
	inStr := os.Getenv("GPIO_TEST_PIN")
	outStr := os.Getenv("GPIO_TEST_PIN_LOOP")
	if gpioDev == "" {
		t.Skip("Please set GPIO_TEST_DEV, GPIO_TEST_PIN to run tests ")
		return "", 0, 0, false, false
	}
	if inStr == "" {
		t.Skip("Please set GPIO_TEST_DEV, GPIO_TEST_PIN to run tests ")
	}
	in, err := strconv.Atoi(inStr)
	if err != nil {
		t.Errorf("GPIO_TEST_PIN pin is unparsable: %s", err)
		return "", 0, 0, false, false
	} else {
		gpioIn = uint32(in)
	}
	hasGpio = true
	if outStr == "" {
		t.Skip("Please set GPIO_TEST_PIN_LOOP and connect it with GPIO_TEST_PIN physically to run tests")
	} else {
		out, err := strconv.Atoi(outStr)
		if err != nil {
			t.Errorf("GPIO_TEST_PIN_LOOP is unparsable: %s", err)
		} else {
			gpioOut = uint32(out)
			hasLoopback = true
		}
	}
	if hasLoopback && gpioIn == gpioOut {
		t.Errorf("test pins can't be same %d|%d", gpioIn, gpioOut)
	}

	return gpioDev, gpioIn, gpioOut, hasGpio, hasLoopback

}

func TestGPIO(t *testing.T) {
	assert := assert.New(t)
	testDevice, testPinInput, testPinOutput, hasGpio, hasLoopback := getEnv(t)
	t.Logf("dev: %s in: %d, out: %d", testDevice, testPinInput, testPinOutput)
	// no test pin provided, skip
	if hasGpio {
		return
	}
	chiper, err := Open(testDevice, "go-test")
	if !assert.Nil(err, "Open should succeed") {
		return
	}
	defer chiper.Close()
	info := chiper.Info()
	assert.NotEqual(
		strings.TrimRight(string(info.Name[:]), "\x00"),
		"",
		"gpiochip should have assert name",
	)

	lineInfo, err := chiper.LineInfo(uint32(testPinInput))
	assert.Nil(err, "LineInfo should succeed")
	assert.NotNil(lineInfo)
	// we dont know the name beforehand to check it
	//assert.Equal(strings.TrimRight(string(lineInfo.Name[:]),"\x00"), "someval")

	readLines, err := chiper.OpenLines(GPIOHANDLE_REQUEST_INPUT, "go-test-in", uint32(testPinInput))
	defer readLines.Close()
	assert.Nil(err, "OpenLines should succeed")
	data, err := readLines.Read()
	assert.Nil(err, "Read should succeed")
	t.Logf("read: %v", data.Values[0])

	if !hasLoopback {
		return
	}

	chiper2, err := Open(testDevice, "go-test-write")
	if !assert.Nil(err, "Open should succeed") {
		return
	}
	defer chiper2.Close()
	writeLines, err := chiper2.OpenLines(GPIOHANDLE_REQUEST_OUTPUT, "go-test-out", uint32(testPinOutput))
	assert.Nil(err, "OpenLines should succeed")
	defer writeLines.Close()
	writeLines.SetBulk(0)

	read1, err := readLines.Read()
	t.Logf("read: %v", read1.Values[0])
	assert.Nil(err, "Read should succeed")

	writeLines.SetBulk(1)
	err = writeLines.Flush()
	assert.Nil(err, "Flush should succeed")

	read2, err := readLines.Read()
	t.Logf("read: %v", read2.Values[0])
	assert.NotEqual(read1.Values[0], read2.Values[0], "line state should change")

	writeLines.SetBulk(0)
	err = writeLines.Flush()
	//time.Sleep(time.Millisecond*100)
	read3, err := readLines.Read()
	t.Logf("read: %v", read3.Values[0])
	assert.Equal(read1.Values[0], read3.Values[0], "line state be back to 0")
}

func TestGPIOEvent(t *testing.T) {
	assert := assert.New(t)
	testDevice, testPinLoopInput, testPinLoopOutput, _, hasLoopback := getEnv(t)
	if !hasLoopback {
		t.Skip("event test requires loopback")
		return
	}
	chiper1, err := Open(testDevice, "go-test-ev-write")
	if !assert.Nil(err) {
		return
	}
	defer chiper1.Close()
	chiper2, err := Open(testDevice, "go-test-ev-read")
	if !assert.Nil(err) {
		return
	}
	defer chiper2.Close()

	writeLines, err := chiper1.OpenLines(GPIOHANDLE_REQUEST_OUTPUT, "go-test-ev-out", testPinLoopOutput)
	if !assert.Nil(err) {
		return
	}
	defer writeLines.Close()
	writeLines.SetBulk(0)
	writeLines.Flush()

	ev, err := chiper2.GetLineEvent(
		testPinLoopInput,
		GPIOHANDLE_REQUEST_INPUT,
		GPIOEVENT_REQUEST_RISING_EDGE,
		"event-read",
	)
	if !assert.Nil(err) {
		return
	}
	defer ev.Close()
	timeStart := time.Now()
	t.Logf("setting event trigger in 10ms")
	go func() {
		time.Sleep(time.Millisecond * 10)
		writeLines.SetBulk(1)
		writeLines.Flush()
	}()
	t.Logf("setting wait for 10s")
	evData, err := ev.Wait(time.Second * 10)
	timeDiff := time.Now().Sub(timeStart)
	t.Logf("triggered after %s", timeDiff)
	t.Logf("event: %+v", evData)
	// timestamp should be present in event
	assert.Greater(evData.Timestamp, uint64(0))
	assert.Equal(int(evData.ID), GPIOEVENT_EVENT_RISING_EDGE)
	// and it should trigger before wait timer is up
	assert.Less(timeDiff.Nanoseconds(), int64(time.Second*9), "should trigger before wait timer is over")

}
