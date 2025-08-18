package tcp_proxy

import (
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
			// thanks to @LucasParsy for his code and patience :)
			if array, ok := ret.([]interface{}); ok {
				result := make([]byte, len(array))
				for i, v := range array {
					if num, ok := v.(float64); ok && num >= 0 && num <= 255 {
						result[i] = byte(num)
					} else {
						log.Error("array element at index %d is not a valid byte value %+v", i, v)
						return nil
					}
				}

				return result
			} else {
				log.Error("error while casting exported value to array of interface: value = %+v error = %+v", ret, err)
			}
		}
	}
	return nil
}
