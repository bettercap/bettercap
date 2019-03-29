package api_rest

import (
	"errors"
	"fmt"
	"time"

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

	mod.Info("loading %s ...", mod.recordFileName)

	start := time.Now()
	if mod.record, err = LoadRecord(mod.recordFileName); err != nil {
		return err
	}
	loadedIn := time.Since(start)

	// we need the api itself up and running
	if !mod.Running() {
		if err := mod.Start(); err != nil {
			return err
		}
	}

	mod.replaying = true
	mod.recording = false

	mod.Info("loaded %d frames in %s, started replaying ...", mod.record.Session.Frames(), loadedIn)

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
