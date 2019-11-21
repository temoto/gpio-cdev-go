package gpio

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type tConfig struct {
	gpioDev     string
	gpioIn      uint32
	gpioOut     uint32
	hasGpio     bool
	hasLoopback bool
	err         error
}

var testConfig = getTestConfig()

// validate and extract test setup
func getTestConfig() (cfg tConfig) {
	cfg.gpioDev = os.Getenv("GPIO_TEST_DEV")
	inStr := os.Getenv("GPIO_TEST_PIN")
	outStr := os.Getenv("GPIO_TEST_PIN_LOOP")
	if cfg.gpioDev == "" || inStr == "" {
		return
	}
	in, err := strconv.Atoi(inStr)
	if err != nil {
		cfg.err = fmt.Errorf("GPIO_TEST_PIN pin is unparsable: %s", err)
	} else {
		cfg.gpioIn = uint32(in)
	}
	cfg.hasGpio = true
	if outStr != "" {
		out, err := strconv.Atoi(outStr)
		if err != nil {
			cfg.err = fmt.Errorf("GPIO_TEST_PIN pin is unparsable: %s", err)
			return cfg
		}
		cfg.gpioOut = uint32(out)
		cfg.hasLoopback = true
	}
	return cfg
}

// test whether we have functional GPIO setup. This is done so we do not need to duplicate the warning messages
func TestGPIOConfig(t *testing.T) {
	if !testConfig.hasGpio {
		t.Skip("Please set GPIO_TEST_DEV, GPIO_TEST_PIN to run tests. See readme.md for details ")
	}
	if !testConfig.hasLoopback {
		t.Skip("Please set GPIO_TEST_PIN_LOOP and connect it with GPIO_TEST_PIN physically to run")
	}
	if testConfig.hasLoopback && testConfig.gpioIn == testConfig.gpioOut {
		t.Errorf("test pins can't be same %d|%d", testConfig.gpioIn, testConfig.gpioOut)
	}

}

func TestGPIO(t *testing.T) {
	assert := assert.New(t)
	if !testConfig.hasGpio {
		t.Skip("gpio test requires gpio")
	}
	chiper, err := Open(testConfig.gpioDev, "go-test")
	if !assert.Nil(err, "Open should succeed") {
		t.FailNow()
	}
	defer chiper.Close()
	info := chiper.Info()
	assert.NotEqual(
		strings.TrimRight(string(info.Name[:]), "\x00"),
		"",
		"gpiochip should have assert name",
	)

	lineInfo, err := chiper.LineInfo(uint32(testConfig.gpioIn))
	assert.Nil(err, "LineInfo should succeed")
	assert.NotNil(lineInfo)
	// we dont know the name beforehand to check it
	//assert.Equal(strings.TrimRight(string(lineInfo.Name[:]),"\x00"), "someval")

	readLines, err := chiper.OpenLines(GPIOHANDLE_REQUEST_INPUT, "go-test-in", uint32(testConfig.gpioIn))
	if !assert.Nil(err, "OpenLines should succeed") {
		t.FailNow()
	}
	defer readLines.Close()
	data, err := readLines.Read()
	assert.Nil(err, "Read should succeed")
	t.Logf("read: %v", data.Values[0])

	if !testConfig.hasLoopback {
		return
	}

	chiper2, err := Open(testConfig.gpioDev, "go-test-write")
	if !assert.Nil(err, "Open should succeed") {
		t.FailNow()
	}
	defer chiper2.Close()

	writeLines, err := chiper2.OpenLines(GPIOHANDLE_REQUEST_OUTPUT, "go-test-out", uint32(testConfig.gpioOut))
	if !assert.Nil(err, "OpenLines should succeed") {
		t.FailNow()
	}
	defer writeLines.Close()
	writeLines.SetBulk(0)

	read1, err := readLines.Read()
	assert.Nil(err, "Read should succeed")
	t.Logf("read: %v", read1.Values[0])

	writeLines.SetBulk(1)
	err = writeLines.Flush()
	assert.Nil(err, "Flush should succeed")

	read2, err := readLines.Read()
	assert.Nil(err, "Read should succeed")
	t.Logf("read: %v", read2.Values[0])
	assert.NotEqual(read1.Values[0], read2.Values[0], "line state should change")

	writeLines.SetBulk(0)
	err = writeLines.Flush()
	assert.Nil(err, "Flush should succeed")
	//time.Sleep(time.Millisecond*100)
	read3, err := readLines.Read()
	assert.Nil(err, "Read should succeed")
	t.Logf("read: %v", read3.Values[0])
	assert.Equal(read1.Values[0], read3.Values[0], "line state be back to 0")
}

func TestGPIOEvent(t *testing.T) {
	assert := assert.New(t)
	if !testConfig.hasLoopback {
		t.Skip("event test requires loopback")
		return
	}
	chiper1, err := Open(testConfig.gpioDev, "go-test-ev-write")
	if !assert.Nil(err) {
		t.FailNow()
	}
	defer chiper1.Close()
	chiper2, err := Open(testConfig.gpioDev, "go-test-ev-read")
	if !assert.Nil(err) {
		t.FailNow()
	}
	defer chiper2.Close()

	writeLines, err := chiper1.OpenLines(GPIOHANDLE_REQUEST_OUTPUT, "go-test-ev-out", testConfig.gpioOut)
	if !assert.Nil(err) {
		t.FailNow()
	}
	defer writeLines.Close()
	writeLines.SetBulk(0)
	writeLines.Flush()

	ev, err := chiper2.GetLineEvent(
		testConfig.gpioIn,
		GPIOHANDLE_REQUEST_INPUT,
		GPIOEVENT_REQUEST_RISING_EDGE,
		"event-read",
	)
	if !assert.Nil(err) {
		t.FailNow()
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
	assert.Nil(err, "Wait should succeed")
	timeDiff := time.Since(timeStart)
	t.Logf("triggered after %s", timeDiff)
	t.Logf("event: %+v", evData)
	// timestamp should be present in event
	assert.Greater(evData.Timestamp, uint64(0))
	assert.Equal(int(evData.ID), GPIOEVENT_EVENT_RISING_EDGE)
	// and it should trigger before wait timer is up
	assert.Less(timeDiff.Nanoseconds(), int64(time.Second*9), "should trigger before wait timer is over")

}
