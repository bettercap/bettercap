function sendMessage(message) {
    var url = 'https://api.telegram.org/bot' + telegramToken +
        '/sendMessage?chat_id=' + telegramChatId +
        '&text=' + http.Encode(message);

    var resp = http.Get(url, {});
    if( resp.Error ) {
        log("error while running sending telegram message: " + resp.Error.Error());
    }
}