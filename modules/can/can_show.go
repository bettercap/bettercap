package can

import (
	"time"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/dustin/go-humanize"
	"github.com/evilsocket/islazy/tui"
)

var (
	AliveTimeInterval      = time.Duration(5) * time.Minute
	PresentTimeInterval    = time.Duration(1) * time.Minute
	JustJoinedTimeInterval = time.Duration(10) * time.Second
)

func (mod *CANModule) getRow(dev *network.CANDevice) []string {
	sinceLastSeen := time.Since(dev.LastSeen)
	seen := dev.LastSeen.Format("15:04:05")

	if sinceLastSeen <= JustJoinedTimeInterval {
		seen = tui.Bold(seen)
	} else if sinceLastSeen > PresentTimeInterval {
		seen = tui.Dim(seen)
	}

	return []string{
		dev.Name,
		dev.Description,
		humanize.Bytes(dev.Read),
		seen,
	}
}

func (mod *CANModule) Show() (err error) {
	devices := mod.Session.CAN.Devices()

	rows := make([][]string, 0)
	for _, dev := range devices {
		rows = append(rows, mod.getRow(dev))
	}

	tui.Table(mod.Session.Events.Stdout, []string{"Name", "Description", "Data", "Seen"}, rows)

	if len(rows) > 0 {
		mod.Session.Refresh()
	}

	return nil
}
