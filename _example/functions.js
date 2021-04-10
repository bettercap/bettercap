var fakeESSID = random.String(16, 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ');
var fakeBSSID = random.Mac()

// uses graph.to_dot and graphviz to generate a png graph
function createGraph(who, where) {
    // generates a .dot file with the graph for this mac
    run('graph.to_dot ' + who);
    // uses graphviz to make a png of it
    run('!dot -Tpng bettergraph.dot > ' + where);
}

function onDeauthentication(event) {
    var data = event.data;

    createGraph(data.address1, '/tmp/graph_deauth.png');

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
    sendPhoto("/tmp/graph_deauth.png");
}

function onNewAP(event){
    var ap = event.data;
    if(ap.hostname == fakeESSID) {
        createGraph(ap.mac, '/tmp/graph_ap.png');

        var message = '🦠 Detected rogue AP:\n\n' +
            // 'Time: ' + event.time + "\n" +
            // 'GPS: lat=' + session.GPS.Latitude + " lon=" + session.GPS.Longitude + " updated_at=" +
            //session.GPS.Updated.String() + "\n\n" +
            'AP: ' + ap.mac + ' (' + ap.vendor + ')';

        // send to telegram bot
        sendMessage(message);
        sendPhoto("/tmp/graph_ap.png");
    }
}

function onHandshake(event){
    var data = event.data;
    var what = 'handshake';

    createGraph(data.station, '/tmp/graph_handshake.png');

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
    sendPhoto("/tmp/graph_handshake.png");
}

function onNewNode(event) {
    var node = event.data;

    if(node.type != 'ssid' && node.type != 'ble_server' && graph.IsConnected(node.type, node.id)) {
        createGraph(node.id, '/tmp/graph_node.png');

        var message = '🖥️  Detected previously unknown ' + node.type + ':\n\n' +
            'Type: ' + node.type + "\n" +
            'MAC: ' + node.id;

        // send to telegram bot
        sendMessage(message);
        sendPhoto("/tmp/graph_node.png");
    }
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