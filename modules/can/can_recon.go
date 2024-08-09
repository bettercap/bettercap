package can

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/str"
	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

type Message struct {
	Frame   can.Frame
	Name    string
	Source  *network.CANDevice
	Signals map[string]string
}

func (mod *CANModule) Configure() error {
	var err error

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, mod.deviceName = mod.StringParam("can.device"); err != nil {
		return err
	} else if err, mod.transport = mod.StringParam("can.transport"); err != nil {
		return err
	} else if mod.transport != "can" && mod.transport != "udp" {
		return errors.New("invalid transport")
	} else if err, mod.dbcPath = mod.StringParam("can.dbc_path"); err != nil {
		return err
	}

	if mod.dbcPath != "" {
		input, err := os.ReadFile(mod.dbcPath)
		if err != nil {
			return fmt.Errorf("can't read %s: %v", mod.dbcPath, err)
		}

		mod.Info("compiling %s ...", mod.dbcPath)

		result, err := dbcCompile(mod.dbcPath, input)
		if err != nil {
			return fmt.Errorf("can't compile %s: %v", mod.dbcPath, err)
		}

		for _, warning := range result.Warnings {
			mod.Warning("%v", warning)
		}

		mod.dbc = result.Database
	} else {
		mod.Warning("no can.dbc_path specified, messages won't be parsed")
	}

	if mod.conn, err = socketcan.DialContext(context.Background(), mod.transport, mod.deviceName); err != nil {
		return err
	}

	mod.recv = socketcan.NewReceiver(mod.conn)
	mod.send = socketcan.NewTransmitter(mod.conn)

	return nil
}

func (mod *CANModule) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("started on %s ...", mod.deviceName)

		for mod.recv.Receive() {
			frame := mod.recv.Frame()
			msg := Message{
				Frame: frame,
			}

			if mod.dbc != nil {
				if message, found := mod.dbc.Message(frame.ID); found {
					msg.Name = message.Name

					sourceName := message.SenderNode
					sourceDesc := ""
					if sender, found := mod.dbc.Node(message.SenderNode); found {
						sourceName = sender.Name
						sourceDesc = sender.Description
					}

					_, msg.Source = mod.Session.CAN.AddIfNew(sourceName, sourceDesc, frame.Data[:])

					msg.Signals = make(map[string]string)

					for _, signal := range message.Signals {
						var value string

						if signal.Length <= 32 && signal.IsFloat {
							value = fmt.Sprintf("%f", signal.UnmarshalFloat(frame.Data))
						} else if signal.Length == 1 {
							value = fmt.Sprintf("%v", signal.UnmarshalBool(frame.Data))
						} else if signal.IsSigned {
							value = fmt.Sprintf("%d", signal.UnmarshalSigned(frame.Data))
						} else {
							value = fmt.Sprintf("%d", signal.UnmarshalUnsigned(frame.Data))
						}

						msg.Signals[signal.Name] = str.Trim(fmt.Sprintf("%s %s", value, signal.Unit))
					}
				}
			}

			mod.Session.Events.Add("can.message", msg)
		}
	})
}

func (mod *CANModule) Stop() error {
	if mod.conn != nil {
		mod.recv.Close()
		mod.conn.Close()
		mod.conn = nil
		mod.recv = nil
		mod.send = nil
		mod.dbc = nil
		mod.dbcPath = ""
	}
	return nil
}
