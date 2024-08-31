package can

import (
	"errors"
	"fmt"
	"net"

	"github.com/bettercap/bettercap/v2/session"
	"github.com/hashicorp/go-bexpr"
	"go.einride.tech/can/pkg/socketcan"
)

type CANModule struct {
	session.SessionModule

	transport  string
	deviceName string
	dumpName   string
	dumpInject bool
	filter     string
	filterExpr *bexpr.Evaluator
	dbc        *DBC
	obd2       *OBD2
	conn       net.Conn
	recv       *socketcan.Receiver
	send       *socketcan.Transmitter
}

func NewCanModule(s *session.Session) *CANModule {
	mod := &CANModule{
		SessionModule: session.NewSessionModule("can", s),
		filter:        "",
		dbc:           &DBC{},
		obd2:          &OBD2{},
		filterExpr:    nil,
		transport:     "can",
		deviceName:    "can0",
		dumpName:      "",
		dumpInject:    false,
	}

	mod.AddParam(session.NewStringParameter("can.device",
		mod.deviceName,
		"",
		"CAN-bus device."))

	mod.AddParam(session.NewStringParameter("can.dump",
		mod.dumpName,
		"",
		"Load CAN traffic from this candump log file."))

	mod.AddParam(session.NewBoolParameter("can.dump.inject",
		fmt.Sprintf("%v", mod.dumpInject),
		"Write CAN traffic read form the candump log file to the selected can.device."))

	mod.AddParam(session.NewStringParameter("can.transport",
		mod.transport,
		"",
		"Network type, can be 'can' for SocketCAN or 'udp'."))

	mod.AddParam(session.NewStringParameter("can.filter",
		"",
		"",
		"Optional boolean expression to select frames to report."))

	mod.AddParam(session.NewBoolParameter("can.parse.obd2",
		"false",
		"Enable built in OBD2 PID parsing."))

	mod.AddHandler(session.NewModuleHandler("can.recon on", "",
		"Start CAN-bus discovery.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("can.recon off", "",
		"Stop CAN-bus discovery.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("can.clear", "",
		"Clear everything collected by the discovery module.",
		func(args []string) error {
			mod.Session.CAN.Clear()
			return nil
		}))

	mod.AddHandler(session.NewModuleHandler("can.show", "",
		"Show a list of detected CAN devices.",
		func(args []string) error {
			return mod.Show()
		}))

	mod.AddHandler(session.NewModuleHandler("can.dbc.load NAME", "can.dbc.load (.+)",
		"Load a DBC file from the list of available ones or from disk.",
		func(args []string) error {
			return mod.dbcLoad(args[0])
		}))

	mod.AddHandler(session.NewModuleHandler("can.inject FRAME_EXPRESSION", `(?i)^can\.inject\s+([a-fA-F0-9#R]+)$`,
		"Parse FRAME_EXPRESSION as 'id#data' and inject it as a CAN frame.",
		func(args []string) error {
			if !mod.Running() {
				return errors.New("can module not running")
			}
			return mod.Inject(args[0])
		}))

	mod.AddHandler(session.NewModuleHandler("can.fuzz ID_OR_NODE_NAME OPTIONAL_SIZE", `(?i)^can\.fuzz\s+([^\s]+)\s*(\d*)$`,
		"If an hexadecimal frame ID is specified, create a randomized version of it and inject it. If a node name is specified, a random message for the given node will be instead used.",
		func(args []string) error {
			if !mod.Running() {
				return errors.New("can module not running")
			}

			return mod.Fuzz(args[0], args[1])
		}))

	return mod
}

func (mod *CANModule) Name() string {
	return "can"
}

func (mod *CANModule) Description() string {
	return "A scanner and frames injection module for CAN-bus."
}

func (mod *CANModule) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}
