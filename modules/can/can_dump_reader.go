package can

import (
	"bufio"
	"context"
	"os"
	"regexp"

	"github.com/evilsocket/islazy/str"
	"go.einride.tech/can"
)

// (1700623093.260875) can0 7E0#0322128C00000000
var dumpLineParser = regexp.MustCompile(`(?m)^\(([\d\.]+)\)\s+([^\s]+)\s+(.+)`)

type dumpEntry struct {
	Time   string
	Device string
	Frame  string
}

func (mod *CANModule) startDumpReader() error {
	mod.Info("loading CAN dump from %s ...", mod.dumpName)

	file, err := os.Open(mod.dumpName)
	if err != nil {
		return err
	}
	defer file.Close()

	entries := make([]dumpEntry, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := str.Trim(scanner.Text())
		if line != "" {
			if m := dumpLineParser.FindStringSubmatch(line); len(m) != 4 {
				mod.Warning("unexpected line: '%s' -> %d matches", line, len(m))
			} else {
				entries = append(entries, dumpEntry{
					Time:   m[1],
					Device: m[2],
					Frame:  m[3],
				})
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	mod.Info("loaded %d entries from candump log", len(entries))

	go func() {
		mod.Info("candump reader started ...")

		for _, entry := range entries {
			frame := can.Frame{}
			if err := frame.UnmarshalString(entry.Frame); err != nil {
				mod.Error("could not unmarshal CAN frame: %v", err)
				continue
			}

			if mod.dumpInject {
				if err := mod.send.TransmitFrame(context.Background(), frame); err != nil {
					mod.Error("could not send CAN frame: %v", err)
				}
			} else {
				mod.onFrame(frame)
			}
		}
	}()

	return nil
}
