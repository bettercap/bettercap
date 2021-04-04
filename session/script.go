package session

import (
	"github.com/evilsocket/islazy/plugin"

	_ "github.com/bettercap/bettercap/js"
)

type Script struct {
	*plugin.Plugin
}

func LoadScript(fileName string, ses *Session) (*Script, error) {
	if p, err := plugin.Load(fileName); err != nil {
		return nil, err
	} else {
		return &Script{
			Plugin:  p,
		}, nil
	}
}