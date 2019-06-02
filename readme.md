# What
Pure Go library to access Linux 4.8+ GPIO chardev interface.

I do this for a single project on particular hardware (32bit ARMv6,7).
If you know how to make this work useful for more people, please take a minute to communicate:
- https://github.com/temoto/gpio-cdev-go/issues/new
- temotor@gmail.com

Ultimate success would be to merge this functionality into periph.io lib.


# Possible issues

- may leak `req.fd` descriptors, TODO test
- 64bit system likely have different ioctl numbers
