package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/juju/errors"
	"github.com/temoto/gpio-cdev-go"
)

func debugfsRead(chip int, line uint32) (byte, error) {
	// oh the joy of breaking rules and leaking fd on purpose
	path := fmt.Sprintf("/sys/kernel/debug/gpio-mockup/gpiochip%d/%d", chip, line)
	fd, err := syscall.Open(path, syscall.O_RDWR|syscall.O_CLOEXEC, 0)
	if err != nil {
		return 0, errors.Annotatef(err, "debugfsRead path=%s open", path)
	}
	defer syscall.Close(fd)

	var bufa [8]byte
	_, err = syscall.Read(fd, bufa[:])
	if err != nil {
		return 0, errors.Annotatef(err, "debugfsRead path=%s read", path)
	}
	switch bufa[0] {
	case '0':
		return 0, nil
	case '1':
		return 1, nil
	default:
		return 0, errors.Errorf("debugfsRead read value=0x%x expected 0/1", bufa[0])
	}
}

func wrapped(chipno int) error {
	devpath := fmt.Sprintf("/dev/gpiochip%d", chipno)
	chip, err := gpio.Open(devpath, "test-program")
	if err != nil {
		return errors.Trace(err)
	}
	chipInfo := chip.Info()
	log.Printf("%s", chipInfo.String())

	{
		rlines := []uint32{0, 1, 2, 3}
		lr, err := chip.OpenLines(gpio.GPIOHANDLE_REQUEST_INPUT, "reader", rlines...)
		if err != nil {
			return errors.Trace(err)
		}
		values, err := lr.Read()
		if err != nil {
			return errors.Trace(err)
		}
		log.Printf("rlines=%v", lr.LineOffsets())
		log.Printf("values=%v", values)
		lr.Close()
	}

	// eh, err := chip.GetLineEvent(uint32(line), 0, gpio.GPIOEVENT_REQUEST_BOTH_EDGES, "waiter")
	// if err != nil {
	// 	return errors.Trace(err)
	// }
	// go func() {
	// 	e, err := eh.Wait()
	// 	if err != nil {
	// 		err = errors.Trace(err)
	// 		log.Fatal(err)
	// 	}
	// 	log.Printf("event=%v", e)
	// }()

	{
		// time.Sleep(1 * time.Millisecond) // enter Wait()
		wlines := []uint32{0, 1, 3}
		lw, err := chip.OpenLines(gpio.GPIOHANDLE_REQUEST_OUTPUT, "writer", wlines...)
		if err != nil {
			return errors.Trace(err)
		}
		log.Printf("wlines=%v", lw.LineOffsets())
		lw.SetBulk(1, 0, 1)
		lw.Flush()

		var vs [4]byte
		vs[0], err = debugfsRead(chipno, 0)
		if err != nil {
			return errors.Annotatef(err, "write check line=%d", 0)
		}
		vs[1], err = debugfsRead(chipno, 1)
		if err != nil {
			return errors.Annotatef(err, "write check line=%d", 1)
		}
		vs[2], err = debugfsRead(chipno, 2)
		if err != nil {
			return errors.Annotatef(err, "write check line=%d", 2)
		}
		vs[3], err = debugfsRead(chipno, 3)
		if err != nil {
			return errors.Annotatef(err, "write check line=%d", 3)
		}
		if !(vs[0] == 1 && vs[1] == 0 && vs[2] == 0 && vs[3] == 1) {
			return errors.Errorf("write check values=%v expected=1 0 0? 1", vs[:])
		}

		time.Sleep(10 * time.Millisecond)
		lw.Close()
	}

	return nil
}

func main() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	cmdline := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	chip := cmdline.Int("chip", 0, "")
	_ = cmdline.Parse(os.Args[1:])
	err := wrapped(*chip)
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	} else {
		log.Printf("success")
	}
}
