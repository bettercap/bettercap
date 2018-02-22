<p align="center">
  <img alt="BetterCap" src="https://raw.githubusercontent.com/bettercap/bettercap/master/media/logo.png" height="140" />
  <p align="center">
    <a href="https://github.com/bettercap/bettercap/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/bettercap/bettercap.svg?style=flat-square"></a>
    <a href="/LICENSE"><img alt="Software License" src="https://img.shields.io/badge/license-GPL3-brightgreen.svg?style=flat-square"></a>
    <a href="https://travis-ci.org/bettercap/bettercap"><img alt="Travis" src="https://img.shields.io/travis/bettercap/bettercap/master.svg?style=flat-square"></a>
    <a href="https://goreportcard.com/report/github.com/bettercap/bettercap"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/bettercap/bettercap?style=flat-square&fuckgithubcache=1"></a>
  </p>
</p>

**bettercap** is a complete reimplementation of bettercap, the Swiss army knife for network attacks and monitoring. It is faster, stabler, smaller, easier to install and to use.

## Using it with Docker

In this repository, BetterCAP is containerized using [Alpine Linux](https://alpinelinux.org/ "") -  a security-oriented, lightweight Linux distribution based on musl libc and busybox. The resulting Docker image is relatively small and easy to manage the dependencies.

To pull latest BetterCAP version of the image:

    $ docker pull bettercap/bettercap

To run:

    $ docker run -it --privileged --net=host bettercap/bettercap -h

## Compilation

Make sure you have a correctly configured **Go >= 1.8** environment, that `$GOPATH/bin` is in `$PATH` and the `libpcap-dev` package installed for your system, then:

    $ go get github.com/bettercap/bettercap

This command will download bettercap, install its dependencies, compile it and move the `bettercap` executable to `$GOPATH/bin`.

Now you can use `sudo bettercap -h` to show the basic command line options and just `sudo bettercap` to start an interactive session on your default network interface.

## Compilation on Windows

Despite Windows support [is not yet 100% complete](https://github.com/bettercap/bettercap/issues/45), it is possible to build bettercap for Microsoft platforms and enjoy 99.99% of the experience. The steps to prepare the building environment are:

- Install [go amd64](https://golang.org/dl/) (add go binaries to your `%PATH%`).
- Install [TDM GCC for amd64](http://tdm-gcc.tdragon.net/download) (add TDM-GCC binaries to your `%PATH%`).
- Also add `TDM-GCC\x86_64-w64-mingw32\bin` to your `%PATH%`.
- Install [winpcap](https://www.winpcap.org/install/default.htm).
- Download [Winpcap developer's pack](https://www.winpcap.org/devel.htm) and extract it to `C:\`.
- Find `wpcap.dll` and `packet.dll` in your PC (typically in `c:\windows\system32`).
- Copy them to some other temp folder or else you'll have to supply Admin privs to the following commands.
- Run `gendef` on those files: `gendef wpcap.dll` and `gendef packet.dll` (obtainable with `MinGW Installation Manager`, package `mingw32-gendef`).

This will generate `.def` files, now we'll generate the static libraries files:

- Run `dlltool --as-flags=--64 -m i386:x86-64 -k --output-lib libwpcap.a --input-def wpcap.def`.
- and `dlltool --as-flags=--64 -m i386:x86-64 -k --output-lib libpacket.a --input-def packet.def`.
- Copy both `libwpcap.a` and `libpacket.a` to `c:\WpdPack\Lib\x64`.

And eventually just `go get github.com/bettercap/bettercap` as you would do on other platforms.

## Cross Compilation

As an example, let's cross compile bettercap for ARM (requires `gcc-arm-linux-gnueabi`, `yacc` and `flex` packages).

Download and cross compile libpcap-1.8.1 for ARM (adjust `PCAPV` to use a different libpcap version):

    cd /tmp
    export PCAPV=1.8.1
    wget https://www.tcpdump.org/release/libpcap-$PCAPV.tar.gz
    tar xvf libpcap-$PCAPV.tar.gz
    cd libpcap-$PCAPV
    export CC=arm-linux-gnueabi-gcc
    ./configure --host=arm-linux --with-pcap=linux
    make

Cross compile bettercap itself:

    cd $GOPATH/src/github.com/bettercap/bettercap
    env CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm CGO_LDFLAGS="-L/tmp/libpcap-$PCAPV" make

## Interactive Mode

If no `-caplet` option is specified, bettercap will start in interactive mode, allowing you to start and stop modules manually, change options and apply new firewall rules on the fly.

To get a grasp of what you can do, type `help` and the general help menu will be shown, you can also have module specific help by using `help module-name` (for instance try with `help net.recon`), to see which modules are running and their configuration at any time, you can use the `active` command.

To print all variables and their values instead, you can use `get *` or `get variable-name` to get a single variable (try with `get gateway.address`), to set a new value you can simply `set variable-name new-value` (a value of `""` will clear the variable contents).

## Prompt

The interactive session prompt can be modified by setting the `$` variable, for instance this:

    set $ something

Will set the prompt to the string `something`, which is not very useful, that is why you can also access variables and use colors/effects by using the proper syntax and operators as you can see from the prompt default configuration `{by}{fw}{cidr} {fb}> {env.iface.ipv4} {reset} {bold}» {reset}`.

| Operator | Description |
| ------------- | ------------- |
| `{bold}` | Set text to bold. |
| `{dim}` | Set dim effect on text. |
| `{r}` | Set text foreground color to red. |
| `{g}` | Set text foreground color to red. |
| `{b}` | Set text foreground color to red. |
| `{y}` | Set text foreground color to red. |
| `{fb}` | Set text foreground color to black. |
| `{fw}` | Set text foreground color to white. |
| `{bdg}` | Set text background color to dark gray. |
| `{br}` | Set text background color to red. |
| `{bg}` | Set text background color to green. |
| `{by}` | Set text background color to yellow. |
| `{blb}` | Set text background color to light blue. |
| `{reset}` | Reset text effects (added by default at the end of the prompt if not specified). |

There are also other operators you can use in order to access specific information about the session.

| Operator | Description |
| ------------- | ------------- |
| `{cidr}` | Selected interface subnet CIDR. |
| `{net.sent}` | Number of bytes being sent by the tool on the network. |
| `{net.sent.human}` | Number of bytes being sent by the tool on the network (human readable form). |
| `{net.errors}` | Number of errors while sending packets. |
| `{net.received}` | Number of bytes being sniffed from the tool on the network. |
| `{net.received.human}` | Number of bytes being sniffed from the tool from the network (human readable form). |
| `{net.packets}` | Number of packets being sniffed by the tool from the network. |

And finally, you can access and use any variable that has been declared in the interactive session using the `{env.NAME-OF-THE-VAR}` operator, for instance, the default prompt is using `{env.iface.ipv4}' that is replaced by the `iface.ipv4` session variable contents ( you can check it using the `get iface.ipv4` command ).

## Caplets

Caplets, or `.cap` files are a powerful way to script bettercap's interactive sessions, think about them as the `.rc` files of Metasploit. You will find updated caplets and modules [in this repository](https://github.com/bettercap/caplets), you're strongly invited to check it out in order to fully understand the features of this tool.

## License

`bettercap` is made with ♥  by [the dev team](https://github.com/orgs/bettercap/people) and it's released under the GPL 3 license.
