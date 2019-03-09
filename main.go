package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"runtime"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/modules"
	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

func main() {
	sess, err := session.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer sess.Close()

	if !tui.Effects() {
		if *sess.Options.NoColors {
			fmt.Printf("\n\nWARNING: Terminal colors have been disabled, view will be very limited.\n\n")
		} else {
			fmt.Printf("\n\nWARNING: This terminal does not support colors, view will be very limited.\n\n")
		}
	}

	if *sess.Options.PrintVersion {
		fmt.Printf("%s v%s (built for %s %s with %s)\n", core.Name, core.Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
		return
	}

	appName := fmt.Sprintf("%s v%s", core.Name, core.Version)
	appBuild := fmt.Sprintf("(built for %s %s with %s)", runtime.GOOS, runtime.GOARCH, runtime.Version())

	fmt.Printf("%s %s [type '%s' for a list of commands]\n\n", tui.Bold(appName), tui.Dim(appBuild), tui.Bold("help"))

	// Load all modules
	modules.LoadModules(sess)

	if err = sess.Start(); err != nil {
		log.Fatal("%s", err)
	}

	// Some modules are enabled by default in order
	// to make the interactive session useful.
	for _, modName := range str.Comma(*sess.Options.AutoStart) {
		if err = sess.Run(modName + " on"); err != nil {
			log.Fatal("error while starting module %s: %s", modName, err)
		}
	}

	// Commands sent with -eval are used to set specific
	// caplet parameters (i.e. arp.spoof.targets) via command
	// line, therefore they need to be executed first otherwise
	// modules might already be started.
	for _, cmd := range session.ParseCommands(*sess.Options.Commands) {
		if err = sess.Run(cmd); err != nil {
			log.Error("error while running '%s': %s", tui.Bold(cmd), tui.Red(err.Error()))
		}
	}

	// Then run the caplet if specified.
	if *sess.Options.Caplet != "" {
		if err = sess.RunCaplet(*sess.Options.Caplet); err != nil {
			log.Error("error while running caplet %s: %s", tui.Bold(*sess.Options.Caplet), tui.Red(err.Error()))
		}
	}

	// Eventually start the interactive session.
	for sess.Active {
		line, err := sess.ReadLine()
		if err != nil {
			if err == io.EOF || err.Error() == "Interrupt" {
				if exitPrompt() {
					sess.Run("exit")
					os.Exit(0)
				}
				continue
			} else {
				log.Fatal("%s", err)
			}
		}

		for _, cmd := range session.ParseCommands(line) {
			if err = sess.Run(cmd); err != nil {
				log.Error("%s", err)
			}
		}
	}
}

func exitPrompt() bool {
	var ans string
	fmt.Printf("Are you sure you want to quit this session? y/n ")
	fmt.Scan(&ans)

	return strings.ToLower(ans) == "y"
}
