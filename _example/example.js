require("config")
require("telegram")

var fakeESSID = random.String(16, 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ');
var fakeBSSID = random.Mac()

function onDeauthentication(event) {
    var data = event.data;

    run('graph.to_dot ' + data.address1);
    run('!dot -Tpng bettergraph.dot > /tmp/graph_deauth.png')

    var message = '🚨 Detected deauthentication frame:\n\n' +
        'Time: ' + event.time + "\n" +
        'GPS: lat=' + session.GPS.Latitude + " lon=" + session.GPS.Longitude + " updated_at=" + session.GPS.Updated.String() + "\n\n" +
        'RSSI: ' + data.rssi + "\n" +
        'Reason: ' + data.reason + "\n" +
        'Address1: ' + data.address1 + "\n" +
        'Address2: ' + data.address2 + "\n" +
        'Address3: ' + data.address3;

    // send to telegram bot
    sendMessage(message);
    sendPhoto("/tmp/graph_deauth.png");
}

function onNewAP(event){
    var ap = event.data;
    if(ap.hostname == fakeESSID) {
        run('graph.to_dot ' + ap.mac);
        run('!dot -Tpng bettergraph.dot > /tmp/graph_ap.png')

        var message = '🚨 Detected possible rogue AP:\n\n' +
            'Time: ' + event.time + "\n" +
            'GPS: lat=' + session.GPS.Latitude + " lon=" + session.GPS.Longitude + " updated_at=" + session.GPS.Updated.String() + "\n\n" +
            'AP: ' + ap.mac + ' (' + ap.vendor + ')';

        // send to telegram bot
        sendMessage(message);
        sendPhoto("/tmp/graph_ap.png");
    }
}

function onHandshake(event){
    var data = event.data;
    var what = 'handshake';

    run('graph.to_dot ' + data.station);
    run('!dot -Tpng bettergraph.dot > /tmp/graph_handshake.png')

    if(data.pmkid != null) {
        what = "RSN PMKID";
    } else if(data.full) {
        what += " (full)";
    } else if(data.half) {
        what += " (half)";
    }

    var message = '💰 Captured ' + what + ':\n\n' +
        'Time: ' + event.time + "\n" +
        'GPS: lat=' + session.GPS.Latitude + " lon=" + session.GPS.Longitude + " updated_at=" + session.GPS.Updated.String() + "\n\n" +
        'Station: ' + data.station + "\n" +
        'AP: ' + data.ap;

    // send to telegram bot
    sendMessage(message);
    sendPhoto("/tmp/graph_handshake.png");
}

function onTick(event) {
    run('wifi.probe ' + fakeBSSID + ' ' + fakeESSID);
}

log("session script loaded, fake AP is " + fakeESSID);

// enable the graph module
run('graph on')

// create an empty ticker so we can run commands every few seconds
run('set ticker.commands ""')
run('set ticker.period 10')
run('ticker on')

// enable recon and probing of new hosts
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
// register for wifi.ap.new events
onEvent('wifi.ap.new', onNewAP);