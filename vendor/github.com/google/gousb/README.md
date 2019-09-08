Introduction
============

[![Build Status][ciimg]][ci]
[![GoDoc][docimg]][doc]
[![Coverage Status][coverimg]][cover]
[![Build status][appveimg]][appveyor]


The gousb package is an attempt at wrapping the libusb library into a Go-like binding.

Supported platforms include:

- linux
- darwin
- windows

This is the release 2.0 of the package [github.com/kylelemons/gousb](https://github.com/kylelemons/gousb).
Its API is not backwards-compatible with version 1.0.
As of 2017-07-13 the 2.0 API is considered stable and 1.0 is deprecated.

[coverimg]: https://coveralls.io/repos/github/google/gousb/badge.svg
[cover]:    https://coveralls.io/github/google/gousb
[ciimg]:    https://travis-ci.org/google/gousb.svg
[ci]:       https://travis-ci.org/google/gousb
[docimg]:   https://godoc.org/github.com/google/gousb?status.svg
[doc]:      https://godoc.org/github.com/google/gousb
[appveimg]: https://ci.appveyor.com/api/projects/status/661qp7x33o3wqe4o?svg=true
[appveyor]: https://ci.appveyor.com/project/zagrodzki/gousb

Documentation
=============
The documentation can be viewed via local godoc or via the excellent [godoc.org](http://godoc.org/):

- [usb](http://godoc.org/github.com/google/gousb)
- [usbid](http://godoc.org/pkg/github.com/google/gousb/usbid)

Installation
============

Dependencies
------------
You must first install [libusb-1.0](https://github.com/libusb/libusb/wiki).  This is pretty straightforward on linux and darwin.  The cgo package should be able to find it if you install it in the default manner or use your distribution's package manager.  How to tell cgo how to find one installed in a non-default place is beyond the scope of this README.

*Note*: If you are installing this on darwin, you will probably need to run `fixlibusb_darwin.sh /usr/local/lib/libusb-1.0/libusb.h` because of an LLVM incompatibility.  It shouldn't break C programs, though I haven't tried it in anger.

Example: lsusb
--------------
The gousb project provides a simple but useful example: lsusb.  This binary will list the USB devices connected to your system and various interesting tidbits about them, their configurations, endpoints, etc.  To install it, run the following command:

    go get -v github.com/google/gousb/lsusb

gousb
-----
If you installed the lsusb example, both libraries below are already installed.

Installing the primary gousb package is really easy:

    go get -v github.com/google/gousb

There is also a `usbid` package that will not be installed by default by this command, but which provides useful information including the human-readable vendor and product codes for detected hardware.  It's not installed by default and not linked into the `gousb` package by default because it adds ~400kb to the resulting binary.  If you want both, they can be installed thus:

    go get -v github.com/google/gousb{,/usbid}

Notes for installation on Windows
---------------------------------

You'll need:

- Gcc - tested on [Win-Builds](http://win-builds.org/) and MSYS/MINGW
- pkg-config - see http://www.mingw.org/wiki/FAQ, "How do I get pkg-config installed?"
- [libusb-1.0](http://sourceforge.net/projects/libusb/files/libusb-1.0/).

Make sure the `libusb-1.0.pc` pkg-config file from libusb was installed
and that the result of the `pkg-config --cflags libusb-1.0` command shows the
correct include path for installed libusb.

After that you can continue with instructions for lsusb/gousb above.

Contributing
============
Contributing to this project will require signing the [Google CLA][cla].
This is the same agreement that is required for contributing to Go itself, so if you have
already filled it out for that, you needn't fill it out again.

[cla]: https://cla.developers.google.com/

