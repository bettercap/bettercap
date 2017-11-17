package core

import "flag"

type Options struct {
	InterfaceName *string
	Caplet        *string
	Debug         *bool
	Silent        *bool
	NoHistory     *bool
}

func ParseOptions() (Options, error) {
	o := Options{
		InterfaceName: flag.String("iface", "", "Network interface to bind to."),
		Caplet:        flag.String("caplet", "", "Read commands from this file instead of goin into interactive mode."),
		Debug:         flag.Bool("debug", false, "Print debug messages."),
		Silent:        flag.Bool("silent", false, "Suppress all logs which are not errors."),
		NoHistory:     flag.Bool("no-history", false, "Disable history file."),
	}

	flag.Parse()

	return o, nil
}
