package modules

import (
	"io/ioutil"
	"net"
	"strings"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"

	"github.com/robertkrimen/otto"
)

type TcpProxyScript struct {
	*ProxyScript
	onDataScript *otto.Script
}

func LoadTcpProxyScriptSource(path, source string, sess *session.Session) (err error, s *TcpProxyScript) {
	err, ps := LoadProxyScriptSource(path, source, sess)
	if err != nil {
		return
	}

	s = &TcpProxyScript{
		ProxyScript:  ps,
		onDataScript: nil,
	}

	if s.hasCallback("onData") {
		s.onDataScript, err = s.VM.Compile("", "onData(from, to, data)")
		if err != nil {
			log.Error("Error while compiling onData callback: %s", err)
			return
		}
	}

	return
}

func LoadTcpProxyScript(path string, sess *session.Session) (err error, s *TcpProxyScript) {
	log.Info("Loading TCP proxy script %s ...", path)

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	return LoadTcpProxyScriptSource(path, string(raw), sess)
}

func (s *TcpProxyScript) doDefines(from, to net.Addr, data []byte) (err error) {
	addrFrom := strings.Split(from.String(), ":")[0]
	addrTo := strings.Split(to.String(), ":")[0]

	if err = s.VM.Set("from", addrFrom); err != nil {
		log.Error("Error while defining from: %s", err)
		return
	} else if err = s.VM.Set("to", addrTo); err != nil {
		log.Error("Error while defining to: %s", err)
		return
	} else if err = s.VM.Set("data", string(data)); err != nil {
		log.Error("Error while defining data: %s", err)
		return
	}
	return
}

func (s *TcpProxyScript) OnData(from, to net.Addr, data []byte) []byte {
	if s.onDataScript != nil {
		s.Lock()
		defer s.Unlock()

		err := s.doDefines(from, to, data)
		if err != nil {
			log.Error("Error while running bootstrap definitions: %s", err)
			return nil
		}

		ret, err := s.VM.Run(s.onDataScript)
		if err != nil {
			log.Error("Error while executing onData callback: %s", err)
			return nil
		}

		if ret.IsUndefined() == false && ret.IsString() {
			return []byte(ret.String())
		}
	}

	return nil
}
