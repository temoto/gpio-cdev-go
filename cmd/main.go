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

	chip, err := gpio.Open(*devpath)
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	chipInfo := chip.Info()
	log.Printf("%s", chipInfo.String())
	evch := make(chan gpio.EventData)
	go func() {
		for e := range evch {
			log.Printf("event %v", e)
		}
	}()
	err = chip.EventLoop(uint32(*line), 0, gpio.GPIOEVENT_REQUEST_BOTH_EDGES, "testing-program", evch, make(chan struct{}))
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
}
