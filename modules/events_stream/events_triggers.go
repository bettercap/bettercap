package events_stream

import (
	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/tui"
)

func (mod *EventsStream) addTrigger(tag string, command string) error {
	if err, id := mod.triggerList.Add(tag, command); err != nil {
		return err
	} else {
		mod.Info("trigger for event %s added with identifier '%s'", tui.Green(tag), tui.Bold(id))
	}
	return nil
}

func (mod *EventsStream) clearTrigger(id string) error {
	if err := mod.triggerList.Del(id); err != nil {
		return err
	}
	return nil
}

func (mod *EventsStream) showTriggers() error {
	colNames := []string{
		"ID",
		"Event",
		"Action",
	}
	rows := [][]string{}

	mod.triggerList.Each(func(id string, t Trigger) {
		rows = append(rows, []string{
			tui.Bold(id),
			tui.Green(t.For),
			t.Action,
		})
	})

	if len(rows) > 0 {
		tui.Table(mod.Session.Events.Stdout, colNames, rows)
		mod.Session.Refresh()
	}

	return nil
}

func (mod *EventsStream) dispatchTriggers(e session.Event) {
	if id, cmds, err, found := mod.triggerList.Dispatch(e); err != nil {
		mod.Error("error while dispatching event %s: %v", e.Tag, err)
	} else if found {
		mod.Debug("running trigger %s (cmds:'%s') for event %v", id, cmds, e)
		for _, cmd := range session.ParseCommands(cmds) {
			if err := mod.Session.Run(cmd); err != nil {
				mod.Error("%s", err.Error())
			}
		}
	}
}
