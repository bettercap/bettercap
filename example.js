const wifiInterface = 'put the wifi interface to put in monitor mode here';
const telegramToken = 'put your telegram bot token here';
const telegramChatId = 'put your telegram chat id here';

function sendMessage(message) {
    var url = 'https://api.telegram.org/bot' + telegramToken +
                '/sendMessage?chat_id=' + telegramChatId +
                '&text=' + http.Encode(message);

    var resp = http.Get(url, {});
    if( resp.Error ) {
        log("error while running sending telegram message: " + resp.Error.Error());
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
onEvent('wifi.deauthentication', function(event){
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
});

// register for wifi.client.handshake events
onEvent('wifi.client.handshake', function(event){
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
});

// register for any event
onEvent(function(event){
    // if endpoint.new or endpoint.lost, clear the screen and show hosts
    if( event.Tag.indexOf('endpoint.') === 0 ) {
        run('clear; net.show');
    }
});
