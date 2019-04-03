package api_rest

import (
	"errors"
	"fmt"
	"time"

	"github.com/bettercap/recording"

	"github.com/evilsocket/islazy/fs"
)

var (
	errNotReplaying = errors.New("not replaying")
)

func (mod *RestAPI) errAlreadyReplaying() error {
	return fmt.Errorf("the module is already replaying a session from %s", mod.recordFileName)
}

func (mod *RestAPI) startReplay(filename string) (err error) {
	if mod.replaying {
		return mod.errAlreadyReplaying()
	} else if mod.recording {
		return mod.errAlreadyRecording()
	} else if mod.recordFileName, err = fs.Expand(filename); err != nil {
		return err
	}

	mod.State.Store("load_progress", 0)
	defer func() {
		mod.State.Store("load_progress", 100.0)
	}()

	mod.loading = true
	defer func() {
		mod.loading = false
	}()

	mod.Info("loading %s ...", mod.recordFileName)

	start := time.Now()
	mod.record, err = recording.Load(mod.recordFileName, func(progress float64, done int, total int) {
		mod.State.Store("load_progress", progress)
	})
	if err != nil {
		return err
	}
	loadedIn := time.Since(start)

	// we need the api itself up and running
	if !mod.Running() {
		if err := mod.Start(); err != nil {
			return err
		}
	}

	mod.recStarted = mod.record.Session.StartedAt()
	mod.recStopped = mod.record.Session.StoppedAt()
	duration := mod.recStopped.Sub(mod.recStarted)
	mod.recTime = int(duration.Seconds())
	mod.replaying = true
	mod.recording = false

	mod.Info("loaded %s of recording (%d frames) started at %s in %s, started replaying ...",
		duration,
		mod.record.Session.Frames(),
		mod.recStarted,
		loadedIn)

	return nil
}

func (mod *RestAPI) stopReplay() error {
	if !mod.replaying {
		return errNotReplaying
	}

	mod.replaying = false

	mod.Info("stopped replaying from %s ...", mod.recordFileName)

	mod.recordFileName = ""

	return nil
}
