package core

import "flag"

type Options struct {
	InterfaceName *string
	Caplet        *string
	Debug         *bool
	Silent        *bool
	NoHistory     *bool
	Commands      *string
	CpuProfile    *string
	MemProfile    *string
}

func ParseOptions() (Options, error) {
	o := Options{
		InterfaceName: flag.String("iface", "", "Network interface to bind to, if empty the default interface will be auto selected."),
		Caplet:        flag.String("caplet", "", "Read commands from this file and execute them in the interactive session."),
		Debug:         flag.Bool("debug", false, "Print debug messages."),
		Silent:        flag.Bool("silent", false, "Suppress all logs which are not errors."),
		NoHistory:     flag.Bool("no-history", false, "Disable interactive session history file."),
		Commands:      flag.String("eval", "", "Run one or more commands separated by ; in the interactive session, used to set variables via command line."),
		CpuProfile:    flag.String("cpu-profile", "", "Write cpu profile `file`."),
		MemProfile:    flag.String("mem-profile", "", "Write memory profile to `file`."),
	}

	flag.Parse()

	return o, nil
}
