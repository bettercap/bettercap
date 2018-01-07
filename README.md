<p align="center">
  <img alt="BetterCap" src="https://www.bettercap.org/assets/logo.png" height="140" />
  <h3 align="center">BetterCAP-NG</h3>
  <p align="center">
    <a href="https://github.com/evilsocket/bettercap-ng/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/evilsocket/bettercap-ng.svg?style=flat-square"></a>
    <a href="/LICENSE"><img alt="Software License" src="https://img.shields.io/badge/license-GPL3-brightgreen.svg?style=flat-square"></a>
    <a href="https://travis-ci.org/evilsocket/bettercap-ng"><img alt="Travis" src="https://img.shields.io/travis/evilsocket/bettercap-ng/master.svg?style=flat-square"></a>
    <a href="https://goreportcard.com/report/github.com/evilsocket/bettercap-ng"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/evilsocket/bettercap-ng?style=flat-square&fuckgithubcache=1"></a>
  </p>
</p>

---

This is a WIP of the new version of [bettercap](https://github.com/evilsocket/bettercap), very alpha, **do not use** ... or do, whatever.

## Compiling

Make sure you have a correctly configured Go >= 1.8 environment and the `libpcap-dev` package installed for your system, then:

    git clone https://github.com/evilsocket/bettercap-ng $GOPATH/src/github.com/evilsocket/bettercap-ng
    cd $GOPATH/src/github.com/evilsocket/bettercap-ng
    make deps
    make

To show the command line options:

    sudo ./bettercap-ng -h

To have an idea of what commands you can use once `bettercap-ng` is started, take a look at the `caplets` scripts folder, each of those commands can be either manually entered during the interactive session, or scripted and loaded from `.cap` files.

## License

`bettercap` and `bettercap-ng` are made with ♥  by [Simone Margaritelli](https://www.evilsocket.net/) and they're released under the GPL 3 license.
