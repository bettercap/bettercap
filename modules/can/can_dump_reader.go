package can

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/evilsocket/islazy/str"
	"go.einride.tech/can"
)

// (1700623093.260875) can0 7E0#0322128C00000000
var dumpLineParser = regexp.MustCompile(`(?m)^\(([\d\.]+)\)\s+([^\s]+)\s+(.+)`)

type dumpEntry struct {
	Time   time.Time
	Device string
	Frame  string
}

func parseTimeval(timeval string) (time.Time, error) {
	parts := strings.Split(timeval, ".")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid timeval format")
	}

	seconds, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid seconds value: %v", err)
	}

	microseconds, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid microseconds value: %v", err)
	}

	return time.Unix(seconds, microseconds*1000), nil
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
		if line := str.Trim(scanner.Text()); line != "" {
			if m := dumpLineParser.FindStringSubmatch(line); len(m) != 4 {
				mod.Warning("unexpected line: '%s' -> %d matches", line, len(m))
			} else if timeval, err := parseTimeval(m[1]); err != nil {
				mod.Warning("can't parse (seconds.microseconds) from line: '%s': %v", line, err)
			} else {
				entries = append(entries, dumpEntry{
					Time:   timeval,
					Device: m[2],
					Frame:  m[3],
				})
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	numEntries := len(entries)
	lastEntry := numEntries - 1

	mod.Info("loaded %d entries from candump log", numEntries)

	go func() {
		mod.Info("candump reader started ...")

		for i, entry := range entries {
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

			// compute delay before the next frame
			if i < lastEntry {
				next := entries[i+1]
				diff := next.Time.Sub(entry.Time)
				time.Sleep(diff)
			}

			if !mod.Running() {
				break
			}
		}
	}()

	return nil
}
