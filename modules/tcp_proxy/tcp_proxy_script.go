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
			return toByteArray(ret)
		}
	}
	return nil
}

func toByteArray(ret interface{}) []byte {
	// Handle different array types that otto.Export() might return
	switch v := ret.(type) {
	case []interface{}:
		// Mixed type array
		result := make([]byte, len(v))
		for i, elem := range v {
			if num, ok := toNumber(elem); ok && num >= 0 && num <= 255 {
				result[i] = byte(num)
			} else {
				log.Error("array element at index %d is not a valid byte value %+v", i, elem)
				return nil
			}
		}
		return result
	case []int64:
		// Array of integers
		result := make([]byte, len(v))
		for i, num := range v {
			if num >= 0 && num <= 255 {
				result[i] = byte(num)
			} else {
				log.Error("array element at index %d is not a valid byte value %d", i, num)
				return nil
			}
		}
		return result
	case []float64:
		// Array of floats
		result := make([]byte, len(v))
		for i, num := range v {
			if num >= 0 && num <= 255 {
				result[i] = byte(num)
			} else {
				log.Error("array element at index %d is not a valid byte value %f", i, num)
				return nil
			}
		}
		return result
	default:
		log.Error("unexpected array type returned from onData: %T, value = %+v", ret, ret)
		return nil
	}
}

// toNumber tries to convert an interface{} to a float64
func toNumber(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	default:
		return 0, false
	}
}
