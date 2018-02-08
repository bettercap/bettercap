package main

import (
	"fmt"
	"io"
	"os"

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
	sess.Register(modules.NewTicker(sess))
	sess.Register(modules.NewMacChanger(sess))
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
	sess.Register(modules.NewWOL(sess))

	if err = sess.Start(); err != nil {
		log.Fatal("%", err)
	}

	if err = sess.Run("events.stream on"); err != nil {
		log.Fatal("%", err)
	} else if err = sess.Run("net.recon on"); err != nil {
		log.Fatal("%", err)
	}

	defer sess.Close()

	if *sess.Options.Caplet != "" {
		if err = sess.RunCaplet(*sess.Options.Caplet); err != nil {
			log.Fatal("%s", err)
		}
	}

	for _, cmd := range session.ParseCommands(*sess.Options.Commands) {
		if err = sess.Run(cmd); err != nil {
			log.Fatal("%s", err)
		}
	}

	for sess.Active {
		fmt.Printf("sess.Active = %v\n", sess.Active)
		line, err := sess.ReadLine()
		if err != nil {
			if err == io.EOF {
				continue
			}
			log.Fatal("%s", err)
		}

		for _, cmd := range session.ParseCommands(line) {
			if err = sess.Run(cmd); err != nil {
				log.Error("%s", err)
			}
		}
	}

	// Windows requires this otherwise the app never exits ...
	os.Exit(0)
}
