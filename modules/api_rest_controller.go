package modules

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write an event to the client.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second
	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
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
	toJSON(w, session.I.BLE)
}

func (api *RestAPI) showBleEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mac := strings.ToLower(params["mac"])
	if dev, found := session.I.BLE.Get(mac); found == true {
		toJSON(w, dev)
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
	toJSON(w, session.I.Lan)
}

func (api *RestAPI) showLanEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mac := strings.ToLower(params["mac"])
	if host, found := session.I.Lan.Get(mac); found == true {
		toJSON(w, host)
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
	toJSON(w, session.I.WiFi)
}

func (api *RestAPI) showWiFiEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mac := strings.ToLower(params["mac"])
	if station, found := session.I.WiFi.Get(mac); found == true {
		toJSON(w, station)
	// cycle through station clients if not a station.
	} else {
		for _, ap := range session.I.WiFi.List() {
			if client, found := ap.Get(mac); found == true {
				toJSON(w, client)
			}
		}
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

func (api *RestAPI) streamEvent(ws *websocket.Conn, event session.Event) error {
	msg, err := json.Marshal(event)
	if err != nil {
		log.Error("Error while creating websocket message: %s", err)
		return err
	}

	ws.SetWriteDeadline(time.Now().Add(writeWait))
	if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
		if !strings.Contains(err.Error(), "closed connection") {
			log.Error("Error while writing websocket message: %s", err)
			return err
		}
	}

	return nil
}

func (api *RestAPI) sendPing(ws *websocket.Conn) error {
	ws.SetWriteDeadline(time.Now().Add(writeWait))
	if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
		log.Error("Error while writing websocket ping message: %s", err)
		return err
	}
	return nil
}

func (api *RestAPI) streamWriter(ws *websocket.Conn, w http.ResponseWriter, r *http.Request) {
	defer ws.Close()

	// first we stream what we already have
	events := session.I.Events.Sorted()
	n := len(events)
	if n > 0 {
		log.Debug("Sending %d events.", n)
		for _, event := range events {
			if err := api.streamEvent(ws, event); err != nil {
				return
			}
		}
	}

	session.I.Events.Clear()

	log.Debug("Listening for events and streaming to ws endpoint ...")

	pingTicker := time.NewTicker(pingPeriod)

	for {
		select {
		case <-pingTicker.C:
			if err := api.sendPing(ws); err != nil {
				return
			}
		case event := <-api.eventListener:
			if err := api.streamEvent(ws, event); err != nil {
				return
			}
		case <-api.quit:
			log.Info("Stopping websocket events streamer ...")
			return
		}
	}
}

func (api *RestAPI) streamReader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Debug("Closing websocket reader.")
			break
		}
	}
}

func (api *RestAPI) showEvents(w http.ResponseWriter, r *http.Request) {
	var err error

	if api.useWebsocket {
		ws, err := api.upgrader.Upgrade(w, r, nil)
		if err != nil {
			if _, ok := err.(websocket.HandshakeError); !ok {
				log.Error("Error while updating api.rest connection to websocket: %s", err)
			}
			return
		}

		log.Debug("Websocket streaming started for %s", r.RemoteAddr)

		go api.streamWriter(ws, w, r)
		api.streamReader(ws)
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
	} else if r.Method == "GET" {
		params := mux.Vars(r)
		if r.URL.String() == "/api/session" {
			api.showSession(w, r)
		} else if strings.HasPrefix(r.URL.String(), "/api/session/ble") {
			if params["mac"] != "" {
				api.showBleEndpoint(w, r)
			} else {
				api.showBle(w, r)
			}
		} else if r.URL.String() == "/api/session/env" {
			api.showEnv(w, r)
		} else if r.URL.String() == "/api/session/gateway" {
			api.showGateway(w, r)
		} else if r.URL.String() == "/api/session/interface" {
			api.showInterface(w, r)
		} else if strings.HasPrefix(r.URL.String(), "/api/session/lan") {
			if params["mac"] != "" {
				api.showLanEndpoint(w, r)
			} else {
				api.showLan(w, r)
			}
		} else if r.URL.String() == "/api/session/options" {
			api.showOptions(w, r)
		} else if r.URL.String() == "/api/session/packets" {
			api.showPackets(w, r)
		} else if r.URL.String() == "/api/session/started-at" {
			api.showStartedAt(w, r)
		} else if strings.HasPrefix(r.URL.String(), "/api/session/wifi") {
			if params["mac"] != "" {
				api.showWiFiEndpoint(w, r)
			} else {
				api.showWiFi(w, r)
			}
		}
	} else if r.Method == "POST" {
		api.runSessionCommand(w, r)
	} else {
		http.Error(w, "Bad Request", 400)
	}
}

func (api *RestAPI) eventsRoute(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)

	if api.checkAuth(r) == false {
		setAuthFailed(w, r)
	} else if r.Method == "GET" {
		api.showEvents(w, r)
	} else if r.Method == "DELETE" {
		api.clearEvents(w, r)
	} else {
		http.Error(w, "Bad Request", 400)
	}
}
