package main

import (
	"flag"
	"log"
	"os"

	"github.com/juju/errors"
	"github.com/temoto/gpio-cdev-go"
)

func main() {
	cmdline := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	devpath := cmdline.String("dev", "/dev/gpiochip0", "")
	line := cmdline.Int("line", 0, "")
	_ = cmdline.Parse(os.Args[1:])

	chip, err := gpio.Open(*devpath, "testing-program")
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	chipInfo := chip.Info()
	log.Printf("%s", chipInfo.String())

	lh, err := chip.OpenLines(gpio.GPIOHANDLE_REQUEST_OUTPUT, "test-out", 7, 8, 9, 10, 20, 198, 199)
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	values, err := lh.Read()
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	log.Printf("lines =%v", lh.LineOffsets())
	log.Printf("values=%v", values)
	lh.Close()

	eh, err := chip.GetLineEvent(uint32(*line), 0, gpio.GPIOEVENT_REQUEST_BOTH_EDGES, "")
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	e, err := eh.Wait()
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	log.Printf("%v", e)
}
