<p align="center">
  <img alt="BetterCap" src="https://raw.githubusercontent.com/bettercap/media/master/logo.png" height="140" />
  <p align="center">
    <a href="https://github.com/bettercap/bettercap/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/bettercap/bettercap.svg?style=flat-square"></a>
    <a href="https://github.com/bettercap/bettercap/blob/master/LICENSE.md"><img alt="Software License" src="https://img.shields.io/badge/license-GPL3-brightgreen.svg?style=flat-square"></a>
    <a href="https://travis-ci.org/bettercap/bettercap"><img alt="Travis" src="https://img.shields.io/travis/bettercap/bettercap/master.svg?style=flat-square"></a>
    <a href="https://goreportcard.com/report/github.com/bettercap/bettercap"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/bettercap/bettercap?style=flat-square&fuckgithubcache=1"></a>
  </p>
</p>

**bettercap** is the Swiss army knife for network attacks and monitoring.

## How to Install

A [precompiled version is available](https://github.com/bettercap/bettercap/releases) for each release, alternatively you can use the latest version of the source code from this repository in order to build your own binary.

Make sure you have a correctly configured **Go >= 1.8** environment, that `$GOPATH/bin` is in `$PATH`, that the `libpcap-dev` and `libnetfilter-queue-dev` (this one is only required on Linux) package installed for your system and then:

    $ go get github.com/bettercap/bettercap

This command will download bettercap, install its dependencies, compile it and move the `bettercap` executable to `$GOPATH/bin`. 

Now you can use `sudo bettercap -h` to show the basic command line options and just `sudo bettercap` to start an 
[interactive session](https://github.com/bettercap/bettercap/wiki/Interactive-Mode) on your default network interface, otherwise you can [load a caplet](https://github.com/bettercap/bettercap/wiki/Caplets) from [the dedicated repository](https://github.com/bettercap/caplets).

## Update

In order to update to an unstable but bleeding edge release from this repository, run the command below:

    $ go get -u github.com/bettercap/bettercap

## Documentation and Examples

The project is documented [in this wiki](https://github.com/bettercap/bettercap/wiki).

## License

`bettercap` is made with ♥  by [the dev team](https://github.com/orgs/bettercap/people) and it's released under the GPL 3 license.
