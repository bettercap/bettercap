package tcp_proxy

import (
	"encoding/json"
	"net"
	"strings"

	"github.com/bettercap/bettercap/v2/log"
	"github.com/bettercap/bettercap/v2/session"

	"github.com/evilsocket/islazy/plugin"

	"github.com/robertkrimen/otto"
)

type TcpProxyScript struct {
	*plugin.Plugin
	doOnData bool
}

func LoadTcpProxyScript(path string, sess *session.Session) (err error, s *TcpProxyScript) {
	log.Info("loading tcp proxy script %s ...", path)

	plug, err := plugin.Load(path)
	if err != nil {
		return
	}

	// define session pointer
	if err = plug.Set("env", sess.Env.Data); err != nil {
		log.Error("error while defining environment: %+v", err)
		return
	}

	// run onLoad if defined
	if plug.HasFunc("onLoad") {
		if _, err = plug.Call("onLoad"); err != nil {
			log.Error("error while executing onLoad callback: %s", "\ntraceback:\n  "+err.(*otto.Error).String())
			return
		}
	}

	s = &TcpProxyScript{
		Plugin:   plug,
		doOnData: plug.HasFunc("onData"),
	}
	return
}

func (s *TcpProxyScript) OnData(from, to net.Addr, data []byte, callback func(call otto.FunctionCall) otto.Value) []byte {
	if s.doOnData {
		addrFrom := strings.Split(from.String(), ":")[0]
		addrTo := strings.Split(to.String(), ":")[0]

		if ret, err := s.Call("onData", addrFrom, addrTo, data, callback); err != nil {
			log.Error("error while executing onData callback: %s", err)
			return nil
		} else if ret != nil {
			return toByteArray(ret)
		}
	}
	return nil
}

func toByteArray(ret interface{}) []byte {
	// this approach is a bit hacky but it handles all cases

	// serialize ret to JSON
	if jsonData, err := json.Marshal(ret); err == nil {
		// attempt to deserialize as []float64
		var back2Array []float64
		if err := json.Unmarshal(jsonData, &back2Array); err == nil {
			result := make([]byte, len(back2Array))
			for i, num := range back2Array {
				if num >= 0 && num <= 255 {
					result[i] = byte(num)
				} else {
					log.Error("array element at index %d is not a valid byte value %d", i, num)
					return nil
				}
			}
			return result
		} else {
			log.Error("failed to deserialize %+v to []float64: %v", ret, err)
		}
	} else {
		log.Error("failed to serialize %+v to JSON: %v", ret, err)
	}

	return nil
}
