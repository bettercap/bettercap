package modules

import (
	"net/http"
	"testing"

	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/session"
)

func getScript(src string) *ProxyScript {
	sess := session.Session{}
	sess.Env = session.NewEnvironment(&sess)

	err, script := LoadProxyScriptSource("", src, &sess)
	if err != nil {
		log.Fatal("%s", err)
	}
	return script
}

func getRequest() *http.Request {
	req, err := http.NewRequest("GET", "http://www.google.com/", nil)
	if err != nil {
		log.Fatal("%s", err)
	}
	return req
}

func BenchmarkOnRequest(b *testing.B) {
	script := getScript("function onRequest(req,res){}")
	req := getRequest()

	for n := 0; n < b.N; n++ {
		script.OnRequest(req)
	}
}
