package api_rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bettercap/recording"

	"github.com/evilsocket/islazy/fs"
)

var (
	errNotRecording = errors.New("not recording")
)

func (mod *RestAPI) errAlreadyRecording() error {
	return fmt.Errorf("the module is already recording to %s", mod.recordFileName)
}

func (mod *RestAPI) recordState() error {
	mod.Session.Lock()
	defer mod.Session.Unlock()

	session := new(bytes.Buffer)
	encoder := json.NewEncoder(session)

	if err := encoder.Encode(mod.Session); err != nil {
		return err
	}

	events := new(bytes.Buffer)
	encoder = json.NewEncoder(events)

	if err := encoder.Encode(mod.getEvents(0)); err != nil {
		return err
	}

	return mod.record.NewState(session.Bytes(), events.Bytes())
}

func (mod *RestAPI) recorder() {
	clock := time.Duration(mod.recClock) * time.Second

	mod.recTime = 0
	mod.recording = true
	mod.replaying = false
	mod.record = recording.New(mod.recordFileName)

	mod.Info("started recording to %s (clock %s) ...", mod.recordFileName, clock)

	mod.recordWait.Add(1)
	defer mod.recordWait.Done()

	tick := time.NewTicker(1 * time.Second)
	lastSampled := time.Time{}

	for range tick.C {
		if !mod.recording {
			break
		}

		mod.recTime++

		if time.Since(lastSampled) >= clock {
			lastSampled = time.Now()
			if err := mod.recordState(); err != nil {
				mod.Error("error while recording: %s", err)
				mod.recording = false
				break
			}
		}
	}

	mod.Info("stopped recording to %s ...", mod.recordFileName)
}

func (mod *RestAPI) startRecording(filename string) (err error) {
	if mod.recording {
		return mod.errAlreadyRecording()
	} else if mod.replaying {
		return mod.errAlreadyReplaying()
	} else if err, mod.recClock = mod.IntParam("api.rest.record.clock"); err != nil {
		return err
	} else if mod.recordFileName, err = fs.Expand(filename); err != nil {
		return err
	}

	// we need the api itself up and running
	if !mod.Running() {
		if err = mod.Start(); err != nil {
			return err
		}
	}

	go mod.recorder()

	return nil
}

func (mod *RestAPI) stopRecording() error {
	if !mod.recording {
		return errNotRecording
	}

	mod.recording = false

	mod.recordWait.Wait()

	err := mod.record.Flush()

	mod.recordFileName = ""
	mod.record = nil

	return err
}
