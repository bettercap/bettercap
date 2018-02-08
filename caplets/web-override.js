// Called before every request is executed, just override the response with 
// our own html web page.
function onRequest(req, res) {
    res.Status      = 200;
    res.ContentType = "text/html";
    res.Headers     = "Connection: close";
    res.Body        =  readFile("caplets/www/index.html");
}
