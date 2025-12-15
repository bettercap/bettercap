package graph

import (
	"io/ioutil"
	"os"
	"time"
)

func (mod *Module) generateDotGraph(bssid string) error {
	mod.wLock.Lock()
	defer mod.wLock.Unlock()

	start := time.Now()
	if err := mod.updateSettings(); err != nil {
		return err
	}

	data, size, discarded, err := mod.db.Dot(bssid,
		mod.settings.dot.layout,
		mod.settings.dot.name,
		mod.settings.disconnected)
	if err != nil {
		return err
	}

	if size > 0 {
		if mod.settings.privacy {
			data = privacyFilter.ReplaceAllString(data, "$1:$2:xx:xx:xx:xx")
		}

		if err := ioutil.WriteFile(mod.settings.dot.output, []byte(data), os.ModePerm); err != nil {
			return err
		} else {
			mod.Info("graph saved to %s in %v (%d edges, %d discarded)",
				mod.settings.dot.output,
				time.Since(start),
				size,
				discarded)
		}
	} else {
		mod.Info("graph is empty")
	}

	return nil
}
