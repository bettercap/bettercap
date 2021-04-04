require("config.js")
require("telegram.js")

function onDeauthentication(event) {
    var data = event.Data;
    var message = '🚨 Detected deauthentication frame:\n\n' +
        'Time: ' + event.Time.String() + "\n" +
        'GPS: lat=' + session.GPS.Latitude + " lon=" + session.GPS.Longitude + " updated_at=" + session.GPS.Updated.String() + "\n\n" +
        'RSSI: ' + data.RSSI + "\n" +
        'Reason: ' + data.Reason + "\n" +
        'Address1: ' + data.Address1 + "\n" +
        'Address2: ' + data.Address2 + "\n" +
        'Address3: ' + data.Address3;

    // send to telegram bot
    sendMessage(message);
}

function onHandshake(event){
    var data = event.Data;
    var what = 'handshake';

    if(data.PMKID != null) {
        what = "RSN PMKID";
    } else if(data.Full) {
        what += " (full)";
    } else if(hand.Half) {
        what += " (half)";
    }

    var message = '💰 Captured ' + what + ':\n\n' +
        'Time: ' + event.Time.String() + "\n" +
        'GPS: lat=' + session.GPS.Latitude + " lon=" + session.GPS.Longitude + " updated_at=" + session.GPS.Updated.String() + "\n\n" +
        'Station: ' + data.Station + "\n" +
        'AP: ' + data.AP;

    // send to telegram bot
    sendMessage(message);
}

function onAnyEvent(event){
    // if endpoint.new or endpoint.lost, clear the screen and show hosts
    if( event.Tag.indexOf('endpoint.') === 0 ) {
        // run('clear; net.show');
    }
}

log("session script loaded");

// enable recon and probing of new hosts
run('net.recon on');
run('net.probe on');
// enable wifi scanning
run('set wifi.interface ' + wifiInterface);
run('wifi.recon on');
// register for wifi.deauthentication events
onEvent('wifi.deauthentication', onDeauthentication);
// register for wifi.client.handshake events
onEvent('wifi.client.handshake', onHandshake);
// register for any event
onEvent(onAnyEvent);
