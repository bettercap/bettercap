require("config")
require("telegram")
require("functions")

log("session script loaded, fake AP is " + fakeESSID);

// enable the graph module so we can extract more historical info
// for each device we see
run('graph on')

// create an empty ticker so we can run commands every few seconds
// this will inject decoy wifi client probes used to detect KARMA
// attacks and in general rogue access points
run('set ticker.commands ""')
run('set ticker.period 10')
run('ticker on')

// enable recon and probing of new hosts on IPv4 and IPv6
run('net.recon on');
run('net.probe on');

// enable wifi scanning
run('set wifi.interface ' + wifiInterface);
run('wifi.recon on');

// send fake client probes every tick
onEvent('tick', onTick);

// register for wifi.deauthentication events
onEvent('wifi.deauthentication', onDeauthentication);

// register for wifi.client.handshake events
onEvent('wifi.client.handshake', onHandshake);

// register for wifi.ap.new events (used to detect rogue APs)
onEvent('wifi.ap.new', onNewAP);

// register for new nodes in the graph
onEvent('graph.node.new', onNewNode);

// register for gateway changes
onEvent('gateway.change', onGatewayChange)