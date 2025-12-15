package can

import (
	"fmt"
	"os"
	"sync"

	"github.com/evilsocket/islazy/str"
	"go.einride.tech/can/pkg/descriptor"
)

type DBC struct {
	sync.RWMutex

	path string
	db   *descriptor.Database
}

func (dbc *DBC) Loaded() bool {
	dbc.RLock()
	defer dbc.RUnlock()

	return dbc.db != nil
}

func (dbc *DBC) LoadFile(mod *CANModule, path string) error {
	input, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("can't read %s: %v", path, err)
	}
	return dbc.LoadData(mod, path, input)
}

func (dbc *DBC) LoadData(mod *CANModule, name string, input []byte) error {
	dbc.Lock()
	defer dbc.Unlock()

	mod.Debug("compiling %s ...", name)

	result, err := dbcCompile(name, input)
	if err != nil {
		return fmt.Errorf("can't compile %s: %v", name, err)
	}

	for _, warning := range result.Warnings {
		mod.Warning("%v", warning)
	}

	dbc.path = name
	dbc.db = result.Database

	mod.Info("%s loaded", name)
	return nil
}

func (dbc *DBC) Parse(mod *CANModule, msg *Message) bool {
	dbc.RLock()
	defer dbc.RUnlock()

	// did we load any DBC database?
	if dbc.db == nil {
		return false
	}

	// if the database contains this message id
	if message, found := dbc.db.Message(msg.Frame.ID); found {
		msg.Name = message.Name

		// find source full info in DBC nodes
		sourceName := message.SenderNode
		sourceDesc := ""
		if sender, found := dbc.db.Node(message.SenderNode); found {
			sourceName = sender.Name
			sourceDesc = sender.Description
		}

		// add CAN source if new
		_, msg.Source = mod.Session.CAN.AddIfNew(sourceName, sourceDesc, msg.Frame.Data[:])

		// parse signals
		for _, signal := range message.Signals {
			var value string

			if signal.Length <= 32 && signal.IsFloat {
				value = fmt.Sprintf("%f", signal.UnmarshalFloat(msg.Frame.Data))
			} else if signal.Length == 1 {
				value = fmt.Sprintf("%v", signal.UnmarshalBool(msg.Frame.Data))
			} else if signal.IsSigned {
				value = fmt.Sprintf("%d", signal.UnmarshalSigned(msg.Frame.Data))
			} else {
				value = fmt.Sprintf("%d", signal.UnmarshalUnsigned(msg.Frame.Data))
			}
			msg.Signals[signal.Name] = str.Trim(fmt.Sprintf("%s %s", value, signal.Unit))
		}

		return true
	}

	return false
}

func (dbc *DBC) MessagesBySender(senderId string) []*descriptor.Message {
	dbc.RLock()
	defer dbc.RUnlock()

	fromSender := make([]*descriptor.Message, 0)

	if dbc.db == nil {
		return fromSender
	}

	for _, msg := range dbc.db.Messages {
		if msg.SenderNode == senderId {
			fromSender = append(fromSender, msg)
		}
	}

	return fromSender
}

func (dbc *DBC) MessageById(frameID uint32) *descriptor.Message {
	dbc.RLock()
	defer dbc.RUnlock()

	if dbc.db == nil {
		return nil
	}

	if message, found := dbc.db.Message(frameID); found {
		return message
	}
	return nil
}

func (dbc *DBC) Messages() []*descriptor.Message {
	dbc.RLock()
	defer dbc.RUnlock()

	if dbc.db == nil {
		return nil
	}

	return dbc.db.Messages
}

func (dbc *DBC) AvailableMessages() []string {
	avail := []string{}
	for _, msg := range dbc.Messages() {
		avail = append(avail, fmt.Sprintf("%d (%s)", msg.ID, msg.Name))
	}
	return avail
}

func (dbc *DBC) Senders() []string {
	dbc.RLock()
	defer dbc.RUnlock()

	senders := make([]string, 0)
	if dbc.db == nil {
		return senders
	}

	uniq := make(map[string]bool)
	for _, msg := range dbc.db.Messages {
		uniq[msg.SenderNode] = true
	}

	for sender := range uniq {
		senders = append(senders, sender)
	}

	return senders
}
