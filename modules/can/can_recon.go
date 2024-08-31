package can

import (
	"errors"

	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/tui"
	"github.com/hashicorp/go-bexpr"
	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

func (mod *CANModule) Configure() error {
	var err error
	var parseOBD bool

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, mod.deviceName = mod.StringParam("can.device"); err != nil {
		return err
	} else if err, mod.dumpName = mod.StringParam("can.dump"); err != nil {
		return err
	} else if err, mod.dumpInject = mod.BoolParam("can.dump.inject"); err != nil {
		return err
	} else if err, parseOBD = mod.BoolParam("can.parse.obd2"); err != nil {
		return err
	} else if err, mod.transport = mod.StringParam("can.transport"); err != nil {
		return err
	} else if mod.transport != "can" && mod.transport != "udp" {
		return errors.New("invalid transport")
	} else if err, mod.filter = mod.StringParam("can.filter"); err != nil {
		return err
	}

	mod.obd2.Enable(parseOBD)

	if mod.filter != "" {
		if mod.filterExpr, err = bexpr.CreateEvaluator(mod.filter); err != nil {
			return err
		}
		mod.Warning("filtering frames with expression %s", tui.Bold(mod.filter))
	}

	if mod.conn, err = socketcan.Dial(mod.transport, mod.deviceName); err != nil {
		return err
	}
	mod.recv = socketcan.NewReceiver(mod.conn)
	mod.send = socketcan.NewTransmitter(mod.conn)

	if mod.dumpName != "" {
		if err = mod.startDumpReader(); err != nil {
			return err
		}
	}

	return nil
}

func (mod *CANModule) isFilteredOut(frame can.Frame, msg Message) bool {
	// if we have an active filter
	if mod.filter != "" {
		if res, err := mod.filterExpr.Evaluate(map[string]interface{}{
			"message": msg,
			"frame":   frame,
		}); err != nil {
			mod.Error("error evaluating '%s': %v", mod.filter, err)
		} else if !res {
			mod.Debug("skipping can message %+v", msg)
			return true
		}
	}

	return false
}

func (mod *CANModule) onFrame(frame can.Frame) {
	msg := NewCanMessage(frame)

	// try to parse with DBC if we have any
	if !mod.dbc.Parse(mod, &msg) {
		// not parsed, if enabled try ODB2
		mod.obd2.Parse(mod, &msg)
	}

	if !mod.isFilteredOut(frame, msg) {
		mod.Session.Events.Add("can.message", msg)
	}
}

const canPrompt = "{br}{fw}{env.can.device} {fb}{reset} {bold}» {reset}"

func (mod *CANModule) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	mod.SetPrompt(canPrompt)

	return mod.SetRunning(true, func() {
		mod.Info("started on %s ...", mod.deviceName)

		for mod.recv.Receive() {
			frame := mod.recv.Frame()
			mod.onFrame(frame)
		}
	})
}

func (mod *CANModule) Stop() error {
	mod.SetPrompt(session.DefaultPrompt)

	return mod.SetRunning(false, func() {
		if mod.conn != nil {
			mod.recv.Close()
			mod.conn.Close()
			mod.conn = nil
			mod.recv = nil
			mod.send = nil
			mod.filter = ""
		}
	})
}
