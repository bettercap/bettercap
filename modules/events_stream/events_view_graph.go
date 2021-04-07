package events_stream

import (
	"fmt"
	"io"

	"github.com/bettercap/bettercap/session"
	"github.com/bettercap/bettercap/modules/graph"

	"github.com/evilsocket/islazy/tui"
)

func (mod *EventsStream) viewGraphEvent(output io.Writer, e session.Event) {
	if e.Tag == "graph.node.new" {
		node := e.Data.(*graph.Node)

		fmt.Fprintf(output, "[%s] [%s] %s %s\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			tui.Yellow(string(node.Type)),
			node.ID)
	} else if e.Tag == "graph.edge.new" {
		data := e.Data.(graph.EdgeEvent)
		fmt.Fprintf(output, "[%s] [%s] %s %s %s %s %s\n",
			e.Time.Format(mod.timeFormat),
			tui.Green(e.Tag),
			tui.Dim(string(data.Left.Type)),
			data.Left.ID,
			tui.Bold(string(data.Edge.Type)),
			tui.Dim(string(data.Right.Type)),
			data.Right.ID)
	}else {
		fmt.Fprintf(output, "[%s] [%s] %v\n", e.Time.Format(mod.timeFormat), tui.Green(e.Tag), e)
	}
}
