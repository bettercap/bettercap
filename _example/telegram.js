function sendMessage(message) {
    log(message);

    var url = 'https://api.telegram.org/bot' + telegramToken +
        '/sendMessage?chat_id=' + telegramChatId +
        '&text=' + http.Encode(message);

    var resp = http.Get(url, {});
    if( resp.Error ) {
        log("error while running sending telegram message: " + resp.Error.Error());
    }
}

function sendPhoto(path) {
    var url = 'https://api.telegram.org/bot' + telegramToken + '/sendPhoto';
    var cmd = 'curl -s -X POST "' + url + '" -F chat_id=' + telegramChatId + ' -F photo="@' + path + '" > /dev/null';
    run("!"+cmd);
}