package core

import (
	"flag"
)

type Options struct {
	InterfaceName string
	Gateway       string
	Caplet        string
	AutoStart     string
	Debug         bool
	Silent        bool
	NoColors      bool
	NoHistory     bool
	PrintVersion  bool
	EnvFile       string
	Commands      string
	CpuProfile    string
	MemProfile    string
	CapletsPath   string
	Script        string
	PcapBufSize   int
}

func ParseOptions() (Options, error) {
	var o Options

	flag.StringVar(&o.InterfaceName, "iface", "", "Network interface to bind to, if empty the default interface will be auto selected.")
	flag.StringVar(&o.Gateway, "gateway-override", "", "Use the provided IP address instead of the default gateway. If not specified or invalid, the default gateway will be used.")
	flag.StringVar(&o.AutoStart, "autostart", "events.stream", "Comma separated list of modules to auto start.")
	flag.StringVar(&o.Caplet, "caplet", "", "Read commands from this file and execute them in the interactive session.")
	flag.BoolVar(&o.Debug, "debug", false, "Print debug messages.")
	flag.BoolVar(&o.PrintVersion, "version", false, "Print the version and exit.")
	flag.BoolVar(&o.Silent, "silent", false, "Suppress all logs which are not errors.")
	flag.BoolVar(&o.NoColors, "no-colors", false, "Disable output color effects.")
	flag.BoolVar(&o.NoHistory, "no-history", false, "Disable interactive session history file.")
	flag.StringVar(&o.EnvFile, "env-file", "", "Load environment variables from this file if found, set to empty to disable environment persistence.")
	flag.StringVar(&o.Commands, "eval", "", "Run one or more commands separated by ; in the interactive session, used to set variables via command line.")
	flag.StringVar(&o.CpuProfile, "cpu-profile", "", "Write cpu profile `file`.")
	flag.StringVar(&o.MemProfile, "mem-profile", "", "Write memory profile to `file`.")
	flag.StringVar(&o.CapletsPath, "caplets-path", "", "Specify an alternative base path for caplets.")
	flag.StringVar(&o.Script, "script", "", "Load a session script.")
	flag.IntVar(&o.PcapBufSize, "pcap-buf-size", -1, "PCAP buffer size, leave to 0 for the default value.")

	flag.Parse()

	return o, nil
}
