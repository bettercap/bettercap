package events_stream

import (
	"fmt"
	"github.com/evilsocket/islazy/zip"
	"os"
	"time"
)

func (mod *EventsStream) doRotation() {
	if mod.output == os.Stdout {
		return
	} else if !mod.rotation.Enabled {
		return
	}

	output, isFile := mod.output.(*os.File)
	if !isFile {
		return
	}

	mod.rotation.Lock()
	defer mod.rotation.Unlock()

	doRotate := false
	if info, err := output.Stat(); err == nil {
		if mod.rotation.How == "size" {
			doRotate = float64(info.Size()) >= float64(mod.rotation.Period*1024*1024)
		} else if mod.rotation.How == "time" {
			doRotate = info.ModTime().Unix()%int64(mod.rotation.Period) == 0
		}
	}

	if doRotate {
		var err error

		name := fmt.Sprintf("%s-%s", mod.outputName, time.Now().Format(mod.rotation.Format))

		if err := output.Close(); err != nil {
			mod.Printf("could not close log for rotation: %s\n", err)
			return
		}

		if err := os.Rename(mod.outputName, name); err != nil {
			mod.Printf("could not rename %s to %s: %s\n", mod.outputName, name, err)
		} else if mod.rotation.Compress {
			zipName := fmt.Sprintf("%s.zip", name)
			if err = zip.Files(zipName, []string{name}); err != nil {
				mod.Printf("error creating %s: %s", zipName, err)
			} else if err = os.Remove(name); err != nil {
				mod.Printf("error deleting %s: %s", name, err)
			}
		}

		mod.output, err = os.OpenFile(mod.outputName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			mod.Printf("could not open %s: %s", mod.outputName, err)
		}
	}
}

