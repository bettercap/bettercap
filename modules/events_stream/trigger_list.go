package events_stream

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/bettercap/bettercap/session"

	"github.com/antchfx/jsonquery"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

var reQueryCapture = regexp.MustCompile(`{{([^}]+)}}`)

type Trigger struct {
	For    string
	Action string
}

type TriggerList struct {
	sync.Mutex
	triggers map[string]Trigger
}

func NewTriggerList() *TriggerList {
	return &TriggerList{
		triggers: make(map[string]Trigger),
	}
}

func (l *TriggerList) Add(tag string, command string) (error, string) {
	l.Lock()
	defer l.Unlock()

	idNum := 0
	command = str.Trim(command)

	for id, t := range l.triggers {
		if t.For == tag {
			if t.Action == command {
				return fmt.Errorf("duplicate: trigger '%s' found for action '%s'", tui.Bold(id), command), ""
			}
			idNum++
		}
	}

	id := fmt.Sprintf("%s-%d", tag, idNum)
	l.triggers[id] = Trigger{
		For:    tag,
		Action: command,
	}

	return nil, id
}

func (l *TriggerList) Del(id string) (err error) {
	l.Lock()
	defer l.Unlock()
	if _, found := l.triggers[id]; found {
		delete(l.triggers, id)
	} else {
		err = fmt.Errorf("trigger '%s' not found", tui.Bold(id))
	}
	return err
}

func (l *TriggerList) Each(cb func(id string, t Trigger)) {
	l.Lock()
	defer l.Unlock()
	for id, t := range l.triggers {
		cb(id, t)
	}
}

func (l *TriggerList) Completer(prefix string) []string {
	ids := []string{}
	l.Each(func(id string, t Trigger) {
		if prefix == "" || strings.HasPrefix(id, prefix) {
			ids = append(ids, id)
		}
	})
	return ids
}

func (l *TriggerList) Dispatch(e session.Event) (ident string, cmd string, err error, found bool) {
	l.Lock()
	defer l.Unlock()

	for id, t := range l.triggers {
		if e.Tag == t.For {
			// this is ugly but it's also the only way to allow
			// the user to do this easily - since each event Data
			// field is an interface and type casting is not possible
			// via golang default text/template system, we transform
			// the field to JSON, parse it again and then allow the
			// user to access it in the command via JSON-Query, example:
			//
			// events.on wifi.client.new "wifi.deauth {{Client\mac}}"
			cmd = t.Action
			found = true
			ident = id
			buf := ([]byte)(nil)
			doc := (*jsonquery.Node)(nil)
			// parse each {EXPR}
			for _, m := range reQueryCapture.FindAllString(t.Action, -1) {
				// parse the event Data field as a JSON objects once
				if doc == nil {
					if buf, err = json.Marshal(e.Data); err != nil {
						err = fmt.Errorf("error while encoding event for trigger %s: %v", tui.Bold(id), err)
						return
					} else if doc, err = jsonquery.Parse(strings.NewReader(string(buf))); err != nil {
						err = fmt.Errorf("error while parsing event for trigger %s: %v", tui.Bold(id), err)
						return
					}
				}
				// {EXPR} -> EXPR
				expr := strings.Trim(m, "{}")
				// use EXPR as a JSON query
				if node := jsonquery.FindOne(doc, expr); node != nil {
					cmd = strings.Replace(cmd, m, node.InnerText(), -1)
				} else {
					err = fmt.Errorf(
						"error while parsing expressionfor trigger %s: '%s' doesn't resolve any object: %v",
						tui.Bold(id),
						expr,
						err,
					)
					return
				}
			}

			return
		}
	}

	return
}
