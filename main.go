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

// Some modules are enabled by default in order
// to make the interactive session useful.
var autoEnableList = []string{
	"events.stream",
	"net.recon",
}

func main() {
	if sess, err = session.New(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if core.NoColors == true {
		fmt.Printf("\n\nWARNING: This terminal does not support colors, view will be very limited.\n\n")
	}

	appName := fmt.Sprintf("%s v%s", core.Name, core.Version)

	fmt.Printf("%s (type '%s' for a list of commands)\n\n", core.Bold(appName), core.Bold("help"))

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
		log.Fatal("%s", err)
	}

	for _, modName := range autoEnableList {
		if err = sess.Run(modName + " on"); err != nil {
			log.Fatal("Error while starting module %s: %", modName, err)
		}
	}

	/*
	 * Commands sent with -eval are used to set specific
	 * caplet parameters (i.e. arp.spoof.targets) via command
	 * line, therefore they need to be executed first otherwise
	 * modules might already be started.
	 */
	for _, cmd := range session.ParseCommands(*sess.Options.Commands) {
		if err = sess.Run(cmd); err != nil {
			log.Fatal("%s", err)
		}
	}

	// Then run the caplet if specified.
	if *sess.Options.Caplet != "" {
		if err = sess.RunCaplet(*sess.Options.Caplet); err != nil {
			log.Fatal("%s", err)
		}
	}

	// Eventually start the interactive session.
	for sess.Active {
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

	sess.Close()

	// Windows requires this otherwise the app never exits ...
	os.Exit(0)
}
