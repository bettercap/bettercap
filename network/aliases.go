package network

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/str"
)

var fileName, _ = fs.Expand("~/bettercap.aliases")

type Aliases struct {
	sync.Mutex

	data map[string]string
}

func LoadAliases() (err error, aliases *Aliases) {
	aliases = &Aliases{
		data: make(map[string]string),
	}

	if fs.Exists(fileName) {
		var file *os.File

		file, err = os.Open(fileName)
		if err != nil {
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, " ", 2)
			mac := str.Trim(parts[0])
			alias := str.Trim(parts[1])
			aliases.data[mac] = alias
		}
	}

	return
}

func (a *Aliases) saveUnlocked() error {
	data := ""
	for mac, alias := range a.data {
		data += fmt.Sprintf("%s %s\n", mac, alias)
	}
	return ioutil.WriteFile(fileName, []byte(data), 0644)
}

func (a *Aliases) Save() error {
	a.Lock()
	defer a.Unlock()

	return a.saveUnlocked()
}

func (a *Aliases) Get(mac string) string {
	a.Lock()
	defer a.Unlock()

	if alias, found := a.data[mac]; found {
		return alias
	}
	return ""
}

func (a *Aliases) Set(mac, alias string) error {
	a.Lock()
	defer a.Unlock()

	if alias != "" {
		a.data[mac] = alias
	} else {
		delete(a.data, mac)
	}

	return a.saveUnlocked()
}

func (a *Aliases) Find(alias string) (mac string, found bool) {
	a.Lock()
	defer a.Unlock()

	for m, a := range a.data {
		if alias == a {
			return m, true
		}
	}

	return "", false
}
