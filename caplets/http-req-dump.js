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

function dumpHeaders(req) {
    log( "> " + BOLD(G("Headers")) );
    for( var i = 0; i < req.Headers.length; i++ ) {
        var header = req.Headers[i];
        log( "  " + B(header.Name) + " : " + DIM(header.Value) );
    }
}

function dumpPlain(req) {
    log( "  > " + BOLD(G("Text")) );

    var body = req.ReadBody();

    log( "   " + Y(body) );
}

function dumpForm(req) {
    log( "  > " + BOLD(G("Form")) );

    var form = req.ParseForm();
    for( var key in form ) {
        log( "   " + B(key) + " : " + Y(form[key]) );
    }
}

function dumpJSON(req) {
    log( "  > " + BOLD(G("JSON")) );

    var body = req.ReadBody();

    // TODO: pretty print json
    log( "   " + Y(body) );
}

function pad(num, size, fill) {
    var s = ""+num;

    while( s.length < size ) {
        s = fill + s;
    }

    return s;
}

function toHex(n) {
    var hex = "0123456789abcdef";
    var h = hex[(0xF0 & n) >> 4] + hex[0x0F & n]; 
    return pad(h, 2, '0');
}

function isPrint(c){
    if( !c ) { return false; }
    var code = c.charCodeAt(0);
    return ( code > 31 ) && ( code < 127 );
}

function dumpHex(raw, linePad) {
    var DataSize = raw.length;
    var Bytes = 16;

    for( var address = 0; address < DataSize; address++ ) {
        var saddr = pad(address, 8, '0');
        var shex  = '';
        var sprint = '';

        var end = address + Bytes;
        for( var i = address; i < end; i++ ) {
            if( i < DataSize ) {
                shex += toHex(raw.charCodeAt(i)) + ' ';
                sprint += isPrint(raw[i]) ? raw[i] : '.';
            } else {
                shex   += '   ';
                sprint += ' ';
            }
        }

        address = end;

        log( linePad + G(saddr) + '  ' + shex + ' ' + sprint );
    }
}

function dumpRaw(req) {
    var body = req.ReadBody();

    log( "  > " + BOLD(G("Body")) + " " + DIM("("+body.length + " bytes)") + "\n" );

    dumpHex(body, "    ");
}

function onRequest(req, res) {
    log( BOLD(req.Client), " > ", B(req.Method), " " + req.Hostname + req.Path + ( req.Query ? "?" + req.Query : '') );

    dumpHeaders(req);

    if( req.ContentType ) {
        log();

        if( req.ContentType.indexOf("text/plain") != -1 ) {
            dumpPlain(req);
        }
        else if( req.ContentType.indexOf("application/x-www-form-urlencoded") != -1 ) {
            dumpForm(req);
        }
        else if( req.ContentType.indexOf("application/json") != -1 ) {
            dumpJSON(req);
        }
        else {
            dumpRaw(req);
        }
    }

    log();
}
