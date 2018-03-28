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

Make sure you have a correctly configured **Go >= 1.8** environment, that `$GOPATH/bin` is in `$PATH`, that the `libpcap-dev` and `libnetfilter-queue-dev` package installed for your system and then:

    $ go get github.com/bettercap/bettercap

This command will download bettercap, install its dependencies, compile it and move the `bettercap` executable to `$GOPATH/bin`. 

Now you can use `sudo bettercap -h` to show the basic command line options and just `sudo bettercap` to start an 
[interactive session](https://github.com/bettercap/bettercap/wiki/Interactive-Mode) on your default network interface, otherwise you can [load a caplet](https://github.com/bettercap/bettercap/wiki/Caplets) from [the dedicated repository](https://github.com/bettercap/caplets).

## Update

In order to update to an unstable but bleeding edge release from this repository, run the command below:

    $ go get -u github.com/bettercap/bettercap

## Documentation and Examples

The project is documented [in this wiki](https://github.com/bettercap/bettercap/wiki).

* **[Known Issues](https://github.com/bettercap/bettercap/wiki/Known-Issues)**
* [Using with Docker](https://github.com/bettercap/bettercap/wiki/Using-with-Docker)
* **Compilation**
  * [on Linux and macOS](https://github.com/bettercap/bettercap/wiki/Compilation-on-Linux-and-macOS)
  * [on Windows](https://github.com/bettercap/bettercap/wiki/Compilation-on-Windows)
  * [on Android](https://github.com/bettercap/bettercap/wiki/Compilation-on-Android)
  * [cross compilation (ARM example)](https://github.com/bettercap/bettercap/wiki/Cross-Compilation-(-ARM-example-))
* [Interactive Mode and Command Line Arguments](https://github.com/bettercap/bettercap/wiki/Interactive-Mode)
* [Changing the Prompt](https://github.com/bettercap/bettercap/wiki/Changing-the-Prompt)
* [Caplets](https://github.com/bettercap/bettercap/wiki/Caplets)

### Modules

* **Core**
  * [events.stream](https://github.com/bettercap/bettercap/wiki/events.stream)
  * [ticker](https://github.com/bettercap/bettercap/wiki/ticker)
  * [api.rest](https://github.com/bettercap/bettercap/wiki/api.rest)
  * [update.check](https://github.com/bettercap/bettercap/wiki/update.check)
* **Bluetooth Low Energy**
  * [ble.recon / enum / write](https://github.com/bettercap/bettercap/wiki/ble)
* **802.11**
  * [wifi.recon / deauth / ap](https://github.com/bettercap/bettercap/wiki/wifi)
* **Ethernet and IP**
  * [net.recon](https://github.com/bettercap/bettercap/wiki/net.recon)
  * [net.probe](https://github.com/bettercap/bettercap/wiki/net.probe)
  * [net.sniff](https://github.com/bettercap/bettercap/wiki/net.sniff)
  * [syn.scan](https://github.com/bettercap/bettercap/wiki/syn.scan)
  * [wake on lan](https://github.com/bettercap/bettercap/wiki/wol)
  * **Spoofers**
    * [arp.spoof](https://github.com/bettercap/bettercap/wiki/arp.spoof)
    * [dhcp6.spoof](https://github.com/bettercap/bettercap/wiki/dhcp6.spoof)
    * [dns.spoof](https://github.com/bettercap/bettercap/wiki/dns.spoof)
   * **Proxies**
     * [tcp.proxy](https://github.com/bettercap/bettercap/wiki/tcp.proxy)
       * [modules](https://github.com/bettercap/bettercap/wiki/tcp.modules)
     * [http.proxy](https://github.com/bettercap/bettercap/wiki/http.proxy)
     * [https.proxy](https://github.com/bettercap/bettercap/wiki/https.proxy)
       * [modules](https://github.com/bettercap/bettercap/wiki/http.modules)
* **Servers**
  * [http.server](https://github.com/bettercap/bettercap/wiki/http.server)
* **Utils**
  * [mac.changer](https://github.com/bettercap/bettercap/wiki/mac.changer)
  * [gps](https://github.com/bettercap/bettercap/wiki/gps)

## License

`bettercap` is made with ♥  by [the dev team](https://github.com/orgs/bettercap/people) and it's released under the GPL 3 license.
