<p align="center">
  <small>Join the project community on our server!</small>
  <br/><br/>
  <a href="https://discord.gg/https://discord.gg/btZpkp45gQ" target="_blank" title="Join our community!">
    <img src="https://dcbadge.limes.pink/api/server/https://discord.gg/btZpkp45gQ"/>
  </a>
</p>
<hr/>

<p align="center">
  <img alt="BetterCap" src="https://raw.githubusercontent.com/bettercap/media/master/logo.png" height="140" />
  <p align="center">
    <a href="https://github.com/bettercap/bettercap/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/bettercap/bettercap.svg?style=flat-square"></a>
    <a href="https://github.com/bettercap/bettercap/blob/master/LICENSE.md"><img alt="Software License" src="https://img.shields.io/badge/license-GPL3-brightgreen.svg?style=flat-square"></a>
    <a href="https://github.com/bettercap/bettercap/actions/workflows/test-on-linux.yml"><img alt="Tests on Linux" src="https://github.com/bettercap/bettercap/actions/workflows/test-on-linux.yml/badge.svg"></a>
    <a href="https://github.com/bettercap/bettercap/actions/workflows/test-on-macos.yml"><img alt="Tests on macOS" src="https://github.com/bettercap/bettercap/actions/workflows/test-on-macos.yml/badge.svg"></a>
    <a href="https://github.com/bettercap/bettercap/actions/workflows/test-on-windows.yml"><img alt="Tests on Windows" src="https://github.com/bettercap/bettercap/actions/workflows/test-on-windows.yml/badge.svg"></a>
    <a href="https://hub.docker.com/r/bettercap/bettercap"><img alt="Docker Hub" src="https://img.shields.io/docker/v/bettercap/bettercap?logo=docker"></a>
    <img src="https://img.shields.io/badge/human-coded-brightgreen?logo=data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiNmZmZmZmYiIHN0cm9rZS13aWR0aD0iMiIgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIiBzdHJva2UtbGluZWpvaW49InJvdW5kIiBjbGFzcz0ibHVjaWRlIGx1Y2lkZS1wZXJzb24tc3RhbmRpbmctaWNvbiBsdWNpZGUtcGVyc29uLXN0YW5kaW5nIj48Y2lyY2xlIGN4PSIxMiIgY3k9IjUiIHI9IjEiLz48cGF0aCBkPSJtOSAyMCAzLTYgMyA2Ii8+PHBhdGggZD0ibTYgOCA2IDIgNi0yIi8+PHBhdGggZD0iTTEyIDEwdjQiLz48L3N2Zz4=" alt="This project is 100% made by humans."/>

  </p>
</p>

bettercap is a powerful, easily extensible and portable framework written in Go which aims to offer to security researchers, red teamers and reverse engineers an **easy to use**, **all-in-one solution** with all the features they might possibly need for performing reconnaissance and attacking [WiFi](https://www.bettercap.org/modules/wifi/) networks, [Bluetooth Low Energy](https://www.bettercap.org/modules/ble/) devices, [CAN-bus](https://www.bettercap.org/modules/canbus/), wireless [HID](https://www.bettercap.org/modules/hid/) devices and [Ethernet](https://www.bettercap.org/modules/ethernet) networks.

![UI](https://raw.githubusercontent.com/bettercap/media/master/ui-events.png)

## Main Features

* **WiFi** networks scanning, [deauthentication attack](https://www.evilsocket.net/2018/07/28/Project-PITA-Writeup-build-a-mini-mass-deauther-using-bettercap-and-a-Raspberry-Pi-Zero-W/), [clientless PMKID association attack](https://www.evilsocket.net/2019/02/13/Pwning-WiFi-networks-with-bettercap-and-the-PMKID-client-less-attack/) and automatic WPA/WPA2/WPA3 client handshakes capture.
* **Bluetooth Low Energy** devices scanning, characteristics enumeration, reading and writing.
* 2.4Ghz wireless devices scanning and **MouseJacking** attacks with over-the-air HID frames injection (with DuckyScript support).
* **CAN-bus** and **DBC** support for decoding, injecting and fuzzing frames.
* Passive and active IP network hosts probing and recon.
* **ARP, DNS, NDP and DHCPv6 spoofers** for MITM attacks on IPv4 and IPv6 based networks.
* **Proxies at packet level, TCP level and HTTP/HTTPS** application level fully scriptable with easy to implement **javascript plugins**.
* A powerful **network sniffer** for **credentials harvesting** which can also be used as a **network protocol fuzzer**.
* A very fast port scanner.
* A powerful [REST API](https://www.bettercap.org/modules/core/api.rest/) with support for asynchronous events notification on websocket to orchestrate your attacks easily.
* **A very convenient [web UI](https://www.bettercap.org/usage/web_ui/).**
* [More!](https://www.bettercap.org/modules/)

## Contributors

<a href="https://github.com/bettercap/bettercap/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=bettercap/bettercap" alt="bettercap project contributors" />
</a>

## License

`bettercap` is made with ♥ and released under the GPL 3 license.

## Stargazers over time

[![Stargazers over time](https://starchart.cc/bettercap/bettercap.svg)](https://starchart.cc/bettercap/bettercap)
