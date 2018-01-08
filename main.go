package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/session"
	"github.com/evilsocket/bettercap-ng/session/modules"
)

var sess *session.Session
var err error

func main() {
	if sess, err = session.New(); err != nil {
		panic(err)
	}

	fmt.Printf("%s v%s\n", core.Name, core.Version)
	fmt.Printf("Build: date=%s os=%s arch=%s\n\n", core.BuildDate, runtime.GOOS, runtime.GOARCH)

	sess.Register(session_modules.NewEventsStream(sess))
	sess.Register(session_modules.NewProber(sess))
	sess.Register(session_modules.NewDiscovery(sess))
	sess.Register(session_modules.NewArpSpoofer(sess))
	sess.Register(session_modules.NewSniffer(sess))
	sess.Register(session_modules.NewHttpProxy(sess))
	sess.Register(session_modules.NewRestAPI(sess))

	if err = sess.Start(); err != nil {
		sess.Events.Log(session.FATAL, "%s", err)
	}

	if err = sess.Run("events.stream on"); err != nil {
		sess.Events.Log(session.FATAL, "%s", err)
	}

	defer sess.Close()

	if *sess.Options.Commands != "" {
		for _, cmd := range strings.Split(*sess.Options.Commands, ";") {
			cmd = strings.Trim(cmd, "\r\n\t ")
			if err = sess.Run(cmd); err != nil {
				sess.Events.Log(session.FATAL, "%s", err)
			}
		}
	}

	if *sess.Options.Caplet != "" {
		if err = sess.RunCaplet(*sess.Options.Caplet); err != nil {
			sess.Events.Log(session.FATAL, "%s", err)
		}
	}

	for sess.Active {
		line, err := sess.ReadLine()
		if err != nil {
			sess.Events.Log(session.FATAL, "%s", err)
		}

		if line == "" || line[0] == '#' {
			continue
		}

		if err = sess.Run(line); err != nil {
			sess.Events.Log(session.ERROR, "%s", err)
		}
	}
}
