package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/modules"
	"github.com/evilsocket/bettercap-ng/session"
)

var sess *session.Session
var err error

func main() {
	if sess, err = session.New(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf(core.Bold("%s v%s\n\n"), core.Name, core.Version)

	sess.Register(modules.NewEventsStream(sess))
	sess.Register(modules.NewProber(sess))
	sess.Register(modules.NewDiscovery(sess))
	sess.Register(modules.NewArpSpoofer(sess))
	sess.Register(modules.NewDHCP6Spoofer(sess))
	sess.Register(modules.NewDNSSpoofer(sess))
	sess.Register(modules.NewSniffer(sess))
	sess.Register(modules.NewHttpServer(sess))
	sess.Register(modules.NewHttpProxy(sess))
	sess.Register(modules.NewHttpsProxy(sess))
	sess.Register(modules.NewRestAPI(sess))

	if err = sess.Start(); err != nil {
		log.Fatal("%", err)
	}

	if err = sess.Run("events.stream on"); err != nil {
		log.Fatal("%", err)
	}

	defer sess.Close()

	if *sess.Options.Commands != "" {
		for _, cmd := range strings.Split(*sess.Options.Commands, ";") {
			cmd = strings.Trim(cmd, "\r\n\t ")
			if err = sess.Run(cmd); err != nil {
				log.Fatal("%s", err)
			}
		}
	}

	if *sess.Options.Caplet != "" {
		if err = sess.RunCaplet(*sess.Options.Caplet); err != nil {
			log.Fatal("%s", err)
		}
	}

	for sess.Active {
		line, err := sess.ReadLine()
		if err != nil {
			log.Fatal("%s", err)
		}

		if line == "" || line[0] == '#' {
			continue
		}

		if err = sess.Run(line); err != nil {
			log.Error("%s", err)
		}
	}
}
