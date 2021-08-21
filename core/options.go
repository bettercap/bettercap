package core

import (
	"flag"
)

type Options struct {
	InterfaceName *string
	Gateway       *string
	Caplet        *string
	AutoStart     *string
	Debug         *bool
	Silent        *bool
	NoColors      *bool
	NoHistory     *bool
	PrintVersion  *bool
	EnvFile       *string
	Commands      *string
	CpuProfile    *string
	MemProfile    *string
	CapletsPath   *string
	Script        *string
	PcapBufSize   *int
}

func ParseOptions() (Options, error) {
	o := Options{
		InterfaceName: flag.String("iface", "", "Network interface to bind to, if empty the default interface will be auto selected."),
		Gateway:       flag.String("gateway-override", "", "Use the provided IP address instead of the default gateway. If not specified or invalid, the default gateway will be used."),
		AutoStart:     flag.String("autostart", "events.stream", "Comma separated list of modules to auto start."),
		Caplet:        flag.String("caplet", "", "Read commands from this file and execute them in the interactive session."),
		Debug:         flag.Bool("debug", false, "Print debug messages."),
		PrintVersion:  flag.Bool("version", false, "Print the version and exit."),
		Silent:        flag.Bool("silent", false, "Suppress all logs which are not errors."),
		NoColors:      flag.Bool("no-colors", false, "Disable output color effects."),
		NoHistory:     flag.Bool("no-history", false, "Disable interactive session history file."),
		EnvFile:       flag.String("env-file", "", "Load environment variables from this file if found, set to empty to disable environment persistence."),
		Commands:      flag.String("eval", "", "Run one or more commands separated by ; in the interactive session, used to set variables via command line."),
		CpuProfile:    flag.String("cpu-profile", "", "Write cpu profile `file`."),
		MemProfile:    flag.String("mem-profile", "", "Write memory profile to `file`."),
		CapletsPath:   flag.String("caplets-path", "", "Specify an alternative base path for caplets."),
		Script:        flag.String("script", "", "Load a session script."),
		PcapBufSize:   flag.Int("pcap-buf-size", -1, "PCAP buffer size, leave to 0 for the default value."),
	}

	flag.Parse()

	return o, nil
}
