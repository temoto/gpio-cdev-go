# What

gpio-cdev-go is pure Go library to access Linux 4.8+ GPIO chardev interface. [![GoDoc](https://godoc.org/github.com/temoto/gpio-cdev-go?status.svg)](https://godoc.org/github.com/temoto/gpio-cdev-go)

I do this for a single project on particular hardware (32bit ARMv6,7).
If you know how to make this work useful for more people, please take a minute to communicate:
- https://github.com/temoto/gpio-cdev-go/issues/new
- temotor@gmail.com

Ultimate success would be to merge this functionality into periph.io lib.


# Possible issues

- may leak `req.fd` descriptors, TODO test
- 64-bit system likely have different ioctl numbers, please try and write me back

# Testing:

* get a 2 free GPIO pins
* jumper them
* set environment variables:
```
export GPIO_DEV_PATH="/dev/gpiochip0"
export GPIO_TEST_PIN="19"
export GPIO_TEST_PIN_LOOP="16"
```


# Flair

[![Build status](https://travis-ci.org/temoto/gpio-cdev-go.svg?branch=master)](https://travis-ci.org/temoto/gpio-cdev-go)
[![Coverage](https://codecov.io/gh/temoto/gpio-cdev-go/branch/master/graph/badge.svg)](https://codecov.io/gh/temoto/gpio-cdev-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/temoto/gpio-cdev-go)](https://goreportcard.com/report/github.com/temoto/gpio-cdev-go)
