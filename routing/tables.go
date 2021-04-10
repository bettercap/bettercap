package routing

import "sync"

var (
	lock  = sync.RWMutex{}
	table = make([]Route, 0)
)

func Table() []Route {
	lock.RLock()
	defer lock.RUnlock()
	return table
}

func Update() ([]Route, error) {
	lock.Lock()
	defer lock.Unlock()
	return update()
}

func Gateway(ip RouteType, device string) (string, error) {
	Update()

	lock.RLock()
	defer lock.RUnlock()

	for _, r := range table {
		if r.Type == ip {
			if device == "" || r.Device == device || r.Device == "" /* windows case */ {
				if r.Default {
					return r.Gateway, nil
				}
			}
		}
	}

	return "", nil
}
