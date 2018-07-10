package modules

import (
	"net/http"
	"testing"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"
)

func getScript(src string) *HttpProxyScript {
	sess := session.Session{}
	sess.Env = session.NewEnvironment(&sess, "")

	err, script := LoadHttpProxyScriptSource("", src, &sess)
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
