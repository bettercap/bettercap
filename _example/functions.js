var fakeESSID = random.String(16, 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ');
var fakeBSSID = random.Mac()

function onDeauthentication(event) {
    var data = event.data;

    var message = '🚨 Detected deauthentication frame:\n\n' +
        // 'Time: ' + event.time + "\n" +
        // 'GPS: lat=' + session.GPS.Latitude + " lon=" + session.GPS.Longitude + " updated_at=" +
        //session.GPS.Updated.String() + "\n\n" +
        'RSSI: ' + data.rssi + "\n" +
        'Reason: ' + data.reason + "\n" +
        'Address1: ' + data.address1 + "\n" +
        'Address2: ' + data.address2 + "\n" +
        'Address3: ' + data.address3 + "\n"
        'AP:\n' + JSON.stringify(data.ap, null, 2);


    // send to telegram bot
    sendMessage(message);
}

function onNewAP(event){
    var ap = event.data;
    if(ap.hostname == fakeESSID) {
        var message = '🦠 Detected rogue AP:\n\n' +
            // 'Time: ' + event.time + "\n" +
            // 'GPS: lat=' + session.GPS.Latitude + " lon=" + session.GPS.Longitude + " updated_at=" +
            //session.GPS.Updated.String() + "\n\n" +
            'AP: ' + ap.mac + ' (' + ap.vendor + ')';

        // send to telegram bot
        sendMessage(message);
    }
}

function onHandshake(event){
    var data = event.data;
    var what = 'handshake';

    if(data.pmkid != null) {
        what = "RSN PMKID";
    } else if(data.full) {
        what += " (full)";
    } else if(data.half) {
        what += " (half)";
    }

    var message = '💰 Captured ' + what + ':\n\n' +
        //'Time: ' + event.time + "\n" +
        //'GPS: lat=' + session.GPS.Latitude + " lon=" + session.GPS.Longitude + " updated_at=" +
        //session.GPS.Updated.String() + "\n\n" +
        'Station: ' + data.station + "\n" +
        'AP: ' + data.ap;

    // send to telegram bot
    sendMessage(message);
}

function onGatewayChange(event) {
    var change = event.data;

    var message = '🚨 Detected ' + change.type + ' gateway change, possible MITM attack:\n\n' +
        'Prev: ' + change.prev.ip + ' (' + change.prev.mac + ")\n" +
        'New: ' + change.new.ip + ' (' + change.new.mac + ")";

    // send to telegram bot
    sendMessage(message);
}

function onTick(event) {
    run('wifi.probe ' + fakeBSSID + ' ' + fakeESSID);
}