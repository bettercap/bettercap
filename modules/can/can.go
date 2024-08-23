package can

import (
	"errors"
	"net"

	"github.com/bettercap/bettercap/v2/session"
	"github.com/hashicorp/go-bexpr"
	"go.einride.tech/can/pkg/socketcan"
)

type CANModule struct {
	session.SessionModule

	deviceName string
	transport  string
	filter     string
	filterExpr *bexpr.Evaluator
	dbc        *DBC
	conn       net.Conn
	recv       *socketcan.Receiver
	send       *socketcan.Transmitter
}

func NewCanModule(s *session.Session) *CANModule {
	mod := &CANModule{
		SessionModule: session.NewSessionModule("can", s),
		filter:        "",
		dbc:           &DBC{},
		filterExpr:    nil,
		transport:     "can",
		deviceName:    "can0",
	}

	mod.AddParam(session.NewStringParameter("can.device",
		mod.deviceName,
		"",
		"CAN-bus device."))

	mod.AddParam(session.NewStringParameter("can.transport",
		mod.transport,
		"",
		"Network type, can be 'can' for SocketCAN or 'udp'."))

	mod.AddParam(session.NewStringParameter("can.filter",
		"",
		"",
		"Optional boolean expression to select frames to report."))

	mod.AddHandler(session.NewModuleHandler("can.recon on", "",
		"Start CAN-bus discovery.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("can.load_dbc PATH", "",
		"Start CAN-bus discovery.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("can.dbc.load NAME", "can.dbc.load (.+)",
		"Load a DBC file from the list of available ones or from disk.",
		func(args []string) error {
			return mod.dbcLoad(args[0])
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

	mod.AddHandler(session.NewModuleHandler("can.inject FRAME_EXPRESSION", `(?i)^can\.inject\s+([a-fA-F0-9#R]+)$`,
		"Parse FRAME_EXPRESSION as 'id#data' and inject it as a CAN frame.",
		func(args []string) error {
			if !mod.Running() {
				return errors.New("can module not running")
			}
			return mod.Inject(args[0])
		}))

	mod.AddHandler(session.NewModuleHandler("can.fuzz ID_OR_NODE_NAME", `(?i)^can\.fuzz\s+(.+)$`,
		"If an integer frame ID is specified, create a randomized version of it and inject it. If a node name is specified, a random message for the given node will be instead used.",
		func(args []string) error {
			if !mod.Running() {
				return errors.New("can module not running")
			}
			return mod.Fuzz(args[0])
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
