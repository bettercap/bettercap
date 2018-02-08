// called when script is loaded
function onLoad() {
    console.log( "PROXY SCRIPT LOADED" );
}

// called before a request is proxied
function onRequest(req, res) {
    if( req.Path == "/test-page" ){
        res.Status      = 200;
        res.ContentType = "text/html";
        res.Headers     = "Server: bettercap-ng\r\n" +
                          "Connection: close";
        res.Body        = "<html>" +
                            "<head>" +
                            "<title>Test Page</title>" +
                            "</head>" +
                            "<body>" +
                                "<div align=\"center\">Hello world from bettercap-ng!</div>" + 
                            "</body>" +
                           "</html>";
    }
}

// called after a request is proxied and there's a response
function onResponse(req, res) {
    if( res.Status == 404 ){
        res.ContentType = "text/html";
        res.Headers     = "Server: bettercap-ng\r\n" +
                          "Connection: close";
        res.Body        = "<html>" +
                            "<head>" +
                            "<title>Test 404 Page</title>" +
                            "</head>" +
                            "<body>" +
                                "<div align=\"center\">Custom 404 from bettercap-ng.</div>" + 
                            "</body>" +
                           "</html>";
    }
}
