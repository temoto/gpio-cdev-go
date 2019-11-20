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

// valudate and extract test setup
func getGpioSetup(t *testing.T) (gpiodev string, gpioin uint32, gpioout uint32) {
	if len(os.Getenv("GPIO_DEV_PATH")) == 0 {
		t.Skip("Please set GPIO_DEV_PATH, GPIO_TEST_PIN to run tests ")
		return "", 0, 0
	}
	if len(os.Getenv("GPIO_TEST_PIN")) == 0 {
		t.Skip("Please set GPIO_DEV_PATH, GPIO_TEST_PIN to run tests ")
		return "",0, 0
	}
	testPinIn, err := strconv.Atoi(os.Getenv("GPIO_TEST_PIN"))
	if (testPinIn < 1 || err != nil) {
		t.Errorf("IO pin is below 1 or unparsable: %d|%s", testPinIn, err)
		return "",0, 0
	}
	if len(os.Getenv("GPIO_TEST_PIN_LOOP")) == 0 {
		t.Skip("Please set GPIO_TEST_PIN_LOOP and connect it with GPIO_TEST_PIN physically to run tests")
	}
	testPinOut, err := strconv.Atoi(os.Getenv("GPIO_TEST_PIN_LOOP"))
	if (testPinOut < 1 || err != nil) {
		t.Errorf("IO loop pin is below 1 or unparsable: %d|%s", testPinOut, err)
		testPinOut = 0
	}
	if testPinIn == testPinOut {
		t.Errorf("test pins can't be same %d|%d", testPinIn, testPinOut)
	}

	return os.Getenv("GPIO_DEV_PATH"),uint32(testPinIn), uint32(testPinOut)

}



func TestGPIO(t *testing.T) {
	a := assert.New(t)
	testDevice, testPin, testPinLoop := getGpioSetup(t)
	fmt.Printf("dev: %s in: %d, out: %d\n", testDevice,testPin,testPinLoop)
	// no test pin provided, skip
	if testPin == 0 { return }
	chiper, err := Open(testDevice, "go-test")
	if !a.Nil(err,"Open should succeed") { return }
	defer chiper.Close()
	info := chiper.Info()
	a.NotEqual(
		strings.TrimRight(string(info.Name[:]),"\x00"),
		"",
		"gpiochip should have a name",
	)


	lineInfo, err := chiper.LineInfo(uint32(testPin))
	a.Nil(err,"LineInfo should succeed")
	a.NotNil(lineInfo)
	// we dont know the name beforehand to check it
	//a.Equal(strings.TrimRight(string(lineInfo.Name[:]),"\x00"), "someval")

	readLines , err := chiper.OpenLines(GPIOHANDLE_REQUEST_INPUT, "go-test-in",uint32(testPin))
	defer readLines.Close()
	a.Nil(err,"OpenLines should succeed")
	data, err := readLines.Read()
	a.Nil(err,"Read should succeed")
	fmt.Printf("read: %v\n",data.Values[0])

	// no loopback pin defined ,exit
	if testPinLoop == 0 {
		return
	}
	chiper2, err := Open(testDevice, "go-test-write")
	if !a.Nil(err,"Open should succeed") { return }
	defer chiper2.Close()
	writeLines, err := chiper2.OpenLines(GPIOHANDLE_REQUEST_OUTPUT,"go-test-out",uint32(testPinLoop))
	a.Nil(err,"OpenLines should succeed")
	defer writeLines.Close()
	writeLines.SetBulk(0)

	read1, err := readLines.Read()
	fmt.Printf("read: %v\n",read1.Values[0])
	a.Nil(err,"Read should succeed")

	writeLines.SetBulk(1)
	err = writeLines.Flush()
	a.Nil(err,"Flush should succeed")

    read2, err := readLines.Read()
	fmt.Printf("read: %v\n",read2.Values[0])
    a.NotEqual(read1.Values[0],read2.Values[0],"line state should change")

	writeLines.SetBulk(0)
	err = writeLines.Flush()
	//time.Sleep(time.Millisecond*100)
	read3, err := readLines.Read()
	fmt.Printf("read: %v\n",read3.Values[0])
    a.Equal(read1.Values[0],read3.Values[0],"line state be back to 0")
}

func TestGPIOEvent(t *testing.T) {
	a := assert.New(t)
	testDevice, testPin, testPinLoop := getGpioSetup(t)
	if testPin == 0 || testPinLoop == 0 {
		t.Skip("event test requires loopback")
		return
	}
	chiper1, err := Open(testDevice, "go-test-ev-write")
	if !a.Nil(err) {return}
	defer chiper1.Close()
	chiper2, err := Open(testDevice, "go-test-ev-read")
	if !a.Nil(err) {return}
	defer chiper2.Close()

	writeLines , err := chiper1.OpenLines(GPIOHANDLE_REQUEST_OUTPUT, "go-test-ev-out",testPinLoop)
	if !a.Nil(err) {return}
	defer writeLines.Close()
	writeLines.SetBulk(0)
	writeLines.Flush()



	ev, err := chiper2.GetLineEvent(
		testPin,
		GPIOHANDLE_REQUEST_INPUT,
		GPIOEVENT_REQUEST_RISING_EDGE,
		"event-read",
	)
	if !a.Nil(err) {return}
	defer ev.Close()
	timeStart := time.Now()
	fmt.Printf("setting event trigger in 10ms\n")
	go func() {
		time.Sleep(time.Millisecond * 10)
		writeLines.SetBulk(1)
		writeLines.Flush()
	} ()
	fmt.Printf("setting wait for 10s\n")
	evData, err := ev.Wait(time.Second * 10)
	timeDiff := time.Now().Sub(timeStart)
	fmt.Printf("triggered after %s\n",timeDiff)
	fmt.Printf("event: %+v\n",evData)
	// timestamp should be present in event
	a.Greater(evData.Timestamp,uint64(0))
	a.Equal(int(evData.ID),GPIOEVENT_EVENT_RISING_EDGE)
	// and it should trigger before wait timer is up
	a.Less(timeDiff.Nanoseconds(),int64(time.Second * 9),"should trigger before wait timer is over")


}