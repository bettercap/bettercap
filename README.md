<p align="center">
  <img alt="BetterCap" src="https://raw.githubusercontent.com/bettercap/media/master/logo.png" height="140" />
  <p align="center">
    <a href="https://github.com/bettercap/bettercap/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/bettercap/bettercap.svg?style=flat-square"></a>
    <a href="https://github.com/bettercap/bettercap/blob/master/LICENSE.md"><img alt="Software License" src="https://img.shields.io/badge/license-GPL3-brightgreen.svg?style=flat-square"></a>
    <a href="https://travis-ci.org/bettercap/bettercap"><img alt="Travis" src="https://img.shields.io/travis/bettercap/bettercap/master.svg?style=flat-square"></a>

  </p>
</p>

bettercap is a powerful, easily extensible and portable framework written in Go which aims to offer to security researchers, red teamers and reverse engineers an **easy to use**, **all-in-one solution** with all the features they might possibly need for performing reconnaissance and attacking [WiFi](https://www.bettercap.org/modules/wifi/) networks, [Bluetooth Low Energy](https://www.bettercap.org/modules/ble/) devices, wireless [HID](https://www.bettercap.org/modules/hid/) devices and [Ethernet](https://www.bettercap.org/modules/ethernet) networks.

![UI](https://raw.githubusercontent.com/bettercap/media/master/ui-events.png)

## Main Features

* **WiFi** networks scanning, [deauthentication attack](https://www.evilsocket.net/2018/07/28/Project-PITA-Writeup-build-a-mini-mass-deauther-using-bettercap-and-a-Raspberry-Pi-Zero-W/), [clientless PMKID association attack](https://www.evilsocket.net/2019/02/13/Pwning-WiFi-networks-with-bettercap-and-the-PMKID-client-less-attack/) and automatic WPA/WPA2 client handshakes capture.
* **Bluetooth Low Energy** devices scanning, characteristics enumeration, reading and writing.
* 2.4Ghz wireless devices scanning and **MouseJacking** attacks with over-the-air HID frames injection (with DuckyScript support).
* Passive and active IP network hosts probing and recon.
* **ARP, DNS, NDP and DHCPv6 spoofers** for MITM attacks on IPv4 and IPv6 based networks.
* **Proxies at packet level, TCP level and HTTP/HTTPS** application level fully scriptable with easy to implement **javascript plugins**.
* A powerful **network sniffer** for **credentials harvesting** which can also be used as a **network protocol fuzzer**.
* A very fast port scanner.
* A powerful [REST API](https://www.bettercap.org/modules/core/api.rest/) with support for asynchronous events notification on websocket to orchestrate your attacks easily.
* **A very convenient [web UI](https://www.bettercap.org/usage/#web-ui).**
* [More!](https://www.bettercap.org/modules/)

## License

`bettercap` is made with ♥  by [the dev team](https://github.com/orgs/bettercap/people) and it's released under the GPL 3 license.

## Stargazers over time

[![Stargazers over time](https://starchart.cc/bettercap/bettercap.svg)](https://starchart.cc/bettercap/bettercap)
