package modules

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"

	"github.com/gorilla/mux"
)

type CommandRequest struct {
	Command string `json:"cmd"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"msg"`
}

func setAuthFailed(w http.ResponseWriter, r *http.Request) {
	log.Warning("Unauthorized authentication attempt from %s", r.RemoteAddr)

	w.Header().Set("WWW-Authenticate", `Basic realm="auth"`)
	w.WriteHeader(401)
	w.Write([]byte("Unauthorized"))
}

func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Add("X-Frame-Options", "DENY")
	w.Header().Add("X-Content-Type-Options", "nosniff")
	w.Header().Add("X-XSS-Protection", "1; mode=block")
	w.Header().Add("Referrer-Policy", "same-origin")
}

func toJSON(w http.ResponseWriter, o interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}

func (api *RestAPI) checkAuth(r *http.Request) bool {
	user, pass, _ := r.BasicAuth()
	// timing attack my ass
	if subtle.ConstantTimeCompare([]byte(user), []byte(api.username)) != 1 {
		return false
	} else if subtle.ConstantTimeCompare([]byte(pass), []byte(api.password)) != 1 {
		return false
	}
	return true
}

func (api *RestAPI) showSession(w http.ResponseWriter, r *http.Request) {
	toJSON(w, session.I)
}

func (api *RestAPI) showBle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mac := strings.ToLower(params["mac"])

	if mac == "" {
		toJSON(w, session.I.BLE)
	} else if dev, found := session.I.BLE.Get(mac); found == true {
		toJSON(w, dev)
	} else {
		http.Error(w, "Not Found", 404)
	}
}

func (api *RestAPI) showEnv(w http.ResponseWriter, r *http.Request) {
	toJSON(w, session.I.Env)
}

func (api *RestAPI) showGateway(w http.ResponseWriter, r *http.Request) {
	toJSON(w, session.I.Gateway)
}

func (api *RestAPI) showInterface(w http.ResponseWriter, r *http.Request) {
	toJSON(w, session.I.Interface)
}

func (api *RestAPI) showLan(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mac := strings.ToLower(params["mac"])

	if mac == "" {
		toJSON(w, session.I.Lan)
	} else if host, found := session.I.Lan.Get(mac); found == true {
		toJSON(w, host)
	} else {
		http.Error(w, "Not Found", 404)
	}
}

func (api *RestAPI) showOptions(w http.ResponseWriter, r *http.Request) {
	toJSON(w, session.I.Options)
}

func (api *RestAPI) showPackets(w http.ResponseWriter, r *http.Request) {
	toJSON(w, session.I.Queue)
}

func (api *RestAPI) showStartedAt(w http.ResponseWriter, r *http.Request) {
	toJSON(w, session.I.StartedAt)
}

func (api *RestAPI) showWiFi(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mac := strings.ToLower(params["mac"])

	if mac == "" {
		toJSON(w, session.I.WiFi)
	} else if station, found := session.I.WiFi.Get(mac); found == true {
		toJSON(w, station)
	} else if client, found := session.I.WiFi.GetClient(mac); found == true {
		toJSON(w, client)
	} else {
		http.Error(w, "Not Found", 404)
	}
}

func (api *RestAPI) runSessionCommand(w http.ResponseWriter, r *http.Request) {
	var err error
	var cmd CommandRequest

	if r.Body == nil {
		http.Error(w, "Bad Request", 400)
	} else if err = json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Bad Request", 400)
	} else if err = session.I.Run(cmd.Command); err != nil {
		http.Error(w, err.Error(), 400)
	} else {
		toJSON(w, APIResponse{Success: true})
	}
}

func (api *RestAPI) showEvents(w http.ResponseWriter, r *http.Request) {
	var err error

	if api.useWebsocket {
		api.startStreamingEvents(w, r)
	} else {
		events := session.I.Events.Sorted()
		nevents := len(events)
		nmax := nevents
		n := nmax

		q := r.URL.Query()
		vals := q["n"]
		if len(vals) > 0 {
			n, err = strconv.Atoi(q["n"][0])
			if err == nil {
				if n > nmax {
					n = nmax
				}
			} else {
				n = nmax
			}
		}

		toJSON(w, events[nevents-n:])
	}
}

func (api *RestAPI) clearEvents(w http.ResponseWriter, r *http.Request) {
	session.I.Events.Clear()
}

func (api *RestAPI) sessionRoute(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)

	if api.checkAuth(r) == false {
		setAuthFailed(w, r)
		return
	} else if r.Method == "POST" {
		api.runSessionCommand(w, r)
		return
	} else if r.Method != "GET" {
		http.Error(w, "Bad Request", 400)
		return
	}

	path := r.URL.String()
	switch {
	case path == "/api/session":
		api.showSession(w, r)

	case path == "/api/session/env":
		api.showEnv(w, r)

	case path == "/api/session/gateway":
		api.showGateway(w, r)

	case path == "/api/session/interface":
		api.showInterface(w, r)

	case strings.HasPrefix(path, "/api/session/lan"):
		api.showLan(w, r)

	case path == "/api/session/options":
		api.showOptions(w, r)

	case path == "/api/session/packets":
		api.showPackets(w, r)

	case path == "/api/session/started-at":
		api.showStartedAt(w, r)

	case strings.HasPrefix(path, "/api/session/ble"):
		api.showBle(w, r)

	case strings.HasPrefix(path, "/api/session/wifi"):
		api.showWiFi(w, r)

	default:
		http.Error(w, "Not Found", 404)
	}
}

func (api *RestAPI) eventsRoute(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)

	if api.checkAuth(r) == false {
		setAuthFailed(w, r)
		return
	}

	if r.Method == "GET" {
		api.showEvents(w, r)
	} else if r.Method == "DELETE" {
		api.clearEvents(w, r)
	} else {
		http.Error(w, "Bad Request", 400)
	}
}
