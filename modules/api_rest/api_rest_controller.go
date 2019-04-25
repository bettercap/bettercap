package api_rest

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

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

func (mod *RestAPI) setAuthFailed(w http.ResponseWriter, r *http.Request) {
	mod.Warning("Unauthorized authentication attempt from %s to %s", r.RemoteAddr, r.URL.String())

	w.Header().Set("WWW-Authenticate", `Basic realm="auth"`)
	w.WriteHeader(401)
	w.Write([]byte("Unauthorized"))
}

func (mod *RestAPI) toJSON(w http.ResponseWriter, o interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(o); err != nil {
		mod.Debug("error while encoding object to JSON: %v", err)
	}
}

func (mod *RestAPI) setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Add("X-Frame-Options", "DENY")
	w.Header().Add("X-Content-Type-Options", "nosniff")
	w.Header().Add("X-XSS-Protection", "1; mode=block")
	w.Header().Add("Referrer-Policy", "same-origin")

	w.Header().Set("Access-Control-Allow-Origin", mod.allowOrigin)
	w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
}

func (mod *RestAPI) checkAuth(r *http.Request) bool {
	if mod.username != "" && mod.password != "" {
		user, pass, _ := r.BasicAuth()
		// timing attack my ass
		if subtle.ConstantTimeCompare([]byte(user), []byte(mod.username)) != 1 {
			return false
		} else if subtle.ConstantTimeCompare([]byte(pass), []byte(mod.password)) != 1 {
			return false
		}
	}
	return true
}

func (mod *RestAPI) patchFrame(buf []byte) (frame map[string]interface{}, err error) {
	// this is ugly but necessary: since we're replaying, the
	// api.rest state object is filled with *old* values (the
	// recorded ones), but the UI needs updated values at least
	// of that in order to understand that a replay is going on
	// and where we are at it. So we need to parse the record
	// back into a session object and update only the api.rest.state
	frame = make(map[string]interface{})

	if err = json.Unmarshal(buf, &frame); err != nil {
		return
	}

	for _, i := range frame["modules"].([]interface{}) {
		m := i.(map[string]interface{})
		if m["name"] == "api.rest" {
			state := m["state"].(map[string]interface{})
			mod.State.Range(func(key interface{}, value interface{}) bool {
				state[key.(string)] = value
				return true
			})
			break
		}
	}

	return
}

func (mod *RestAPI) showSession(w http.ResponseWriter, r *http.Request) {
	if mod.replaying {
		if !mod.record.Session.Over() {
			from := mod.record.Session.Index() - 1
			q := r.URL.Query()
			vals := q["from"]
			if len(vals) > 0 {
				if n, err := strconv.Atoi(vals[0]); err == nil {
					from = n
				}
			}
			mod.record.Session.SetFrom(from)

			mod.Debug("replaying session %d of %d from %s",
				mod.record.Session.Index(),
				mod.record.Session.Frames(),
				mod.recordFileName)

			mod.State.Store("rec_frames", mod.record.Session.Frames())
			mod.State.Store("rec_cur_frame", mod.record.Session.Index())

			buf := mod.record.Session.Next()
			if frame, err := mod.patchFrame(buf); err != nil {
				mod.Error("%v", err)
			} else {
				mod.toJSON(w, frame)
				return
			}
		} else {
			mod.stopReplay()
		}
	}

	mod.toJSON(w, mod.Session)
}

func (mod *RestAPI) showBLE(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mac := strings.ToLower(params["mac"])

	if mac == "" {
		mod.toJSON(w, mod.Session.BLE)
	} else if dev, found := mod.Session.BLE.Get(mac); found {
		mod.toJSON(w, dev)
	} else {
		http.Error(w, "Not Found", 404)
	}
}

func (mod *RestAPI) showHID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mac := strings.ToLower(params["mac"])

	if mac == "" {
		mod.toJSON(w, mod.Session.HID)
	} else if dev, found := mod.Session.HID.Get(mac); found {
		mod.toJSON(w, dev)
	} else {
		http.Error(w, "Not Found", 404)
	}
}

func (mod *RestAPI) showEnv(w http.ResponseWriter, r *http.Request) {
	mod.toJSON(w, mod.Session.Env)
}

func (mod *RestAPI) showGateway(w http.ResponseWriter, r *http.Request) {
	mod.toJSON(w, mod.Session.Gateway)
}

func (mod *RestAPI) showInterface(w http.ResponseWriter, r *http.Request) {
	mod.toJSON(w, mod.Session.Interface)
}

func (mod *RestAPI) showModules(w http.ResponseWriter, r *http.Request) {
	mod.toJSON(w, mod.Session.Modules)
}

func (mod *RestAPI) showLAN(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mac := strings.ToLower(params["mac"])

	if mac == "" {
		mod.toJSON(w, mod.Session.Lan)
	} else if host, found := mod.Session.Lan.Get(mac); found {
		mod.toJSON(w, host)
	} else {
		http.Error(w, "Not Found", 404)
	}
}

func (mod *RestAPI) showOptions(w http.ResponseWriter, r *http.Request) {
	mod.toJSON(w, mod.Session.Options)
}

func (mod *RestAPI) showPackets(w http.ResponseWriter, r *http.Request) {
	mod.toJSON(w, mod.Session.Queue)
}

func (mod *RestAPI) showStartedAt(w http.ResponseWriter, r *http.Request) {
	mod.toJSON(w, mod.Session.StartedAt)
}

func (mod *RestAPI) showWiFi(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mac := strings.ToLower(params["mac"])

	if mac == "" {
		mod.toJSON(w, mod.Session.WiFi)
	} else if station, found := mod.Session.WiFi.Get(mac); found {
		mod.toJSON(w, station)
	} else if client, found := mod.Session.WiFi.GetClient(mac); found {
		mod.toJSON(w, client)
	} else {
		http.Error(w, "Not Found", 404)
	}
}

func (mod *RestAPI) runSessionCommand(w http.ResponseWriter, r *http.Request) {
	var err error
	var cmd CommandRequest

	if r.Body == nil {
		http.Error(w, "Bad Request", 400)
	} else if err = json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Bad Request", 400)
	}

	for _, aCommand := range session.ParseCommands(cmd.Command) {
		if err = mod.Session.Run(aCommand); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}

	mod.toJSON(w, APIResponse{Success: true})
}

func (mod *RestAPI) getEvents(limit int) []session.Event {
	events := make([]session.Event, 0)
	for _, e := range mod.Session.Events.Sorted() {
		if mod.Session.EventsIgnoreList.Ignored(e) == false {
			events = append(events, e)
		}
	}

	nevents := len(events)
	nmax := nevents
	n := nmax

	if limit > 0 && limit < nmax {
		n = limit
	}

	return events[nevents-n:]
}

func (mod *RestAPI) showEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if mod.replaying {
		if !mod.record.Events.Over() {
			from := mod.record.Events.Index() - 1
			vals := q["from"]
			if len(vals) > 0 {
				if n, err := strconv.Atoi(vals[0]); err == nil {
					from = n
				}
			}
			mod.record.Events.SetFrom(from)

			mod.Debug("replaying events %d of %d from %s",
				mod.record.Events.Index(),
				mod.record.Events.Frames(),
				mod.recordFileName)

			buf := mod.record.Events.Next()
			if _, err := w.Write(buf); err != nil {
				mod.Error("%v", err)
			} else {
				return
			}
		} else {
			mod.stopReplay()
		}
	}

	if mod.useWebsocket {
		mod.startStreamingEvents(w, r)
	} else {
		vals := q["n"]
		limit := 0
		if len(vals) > 0 {
			if n, err := strconv.Atoi(q["n"][0]); err == nil {
				limit = n
			}
		}

		mod.toJSON(w, mod.getEvents(limit))
	}
}

func (mod *RestAPI) clearEvents(w http.ResponseWriter, r *http.Request) {
	mod.Session.Events.Clear()
}

func (mod *RestAPI) corsRoute(w http.ResponseWriter, r *http.Request) {
	mod.setSecurityHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

func (mod *RestAPI) sessionRoute(w http.ResponseWriter, r *http.Request) {
	mod.setSecurityHeaders(w)

	if !mod.checkAuth(r) {
		mod.setAuthFailed(w, r)
		return
	} else if r.Method == "POST" {
		mod.runSessionCommand(w, r)
		return
	} else if r.Method != "GET" {
		http.Error(w, "Bad Request", 400)
		return
	}

	mod.Session.Lock()
	defer mod.Session.Unlock()

	path := r.URL.Path
	switch {
	case path == "/api/session":
		mod.showSession(w, r)

	case path == "/api/session/env":
		mod.showEnv(w, r)

	case path == "/api/session/gateway":
		mod.showGateway(w, r)

	case path == "/api/session/interface":
		mod.showInterface(w, r)

	case strings.HasPrefix(path, "/api/session/modules"):
		mod.showModules(w, r)

	case strings.HasPrefix(path, "/api/session/lan"):
		mod.showLAN(w, r)

	case path == "/api/session/options":
		mod.showOptions(w, r)

	case path == "/api/session/packets":
		mod.showPackets(w, r)

	case path == "/api/session/started-at":
		mod.showStartedAt(w, r)

	case strings.HasPrefix(path, "/api/session/ble"):
		mod.showBLE(w, r)

	case strings.HasPrefix(path, "/api/session/hid"):
		mod.showHID(w, r)

	case strings.HasPrefix(path, "/api/session/wifi"):
		mod.showWiFi(w, r)

	default:
		http.Error(w, "Not Found", 404)
	}
}

func (mod *RestAPI) readFile(fileName string, w http.ResponseWriter, r *http.Request) {
	fp, err := os.Open(fileName)
	if err != nil {
		msg := fmt.Sprintf("could not open %s for reading: %s", fileName, err)
		mod.Debug(msg)
		http.Error(w, msg, 404)
		return
	}
	defer fp.Close()

	w.Header().Set("Content-type", "application/octet-stream")

	io.Copy(w, fp)
}

func (mod *RestAPI) writeFile(fileName string, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("invalid file upload: %s", err)
		mod.Warning(msg)
		http.Error(w, msg, 404)
		return
	}

	err = ioutil.WriteFile(fileName, data, 0666)
	if err != nil {
		msg := fmt.Sprintf("can't write to %s: %s", fileName, err)
		mod.Warning(msg)
		http.Error(w, msg, 404)
		return
	}

	mod.toJSON(w, APIResponse{
		Success: true,
		Message: fmt.Sprintf("%s created", fileName),
	})
}

func (mod *RestAPI) eventsRoute(w http.ResponseWriter, r *http.Request) {
	mod.setSecurityHeaders(w)

	if !mod.checkAuth(r) {
		mod.setAuthFailed(w, r)
		return
	}

	if r.Method == "GET" {
		mod.showEvents(w, r)
	} else if r.Method == "DELETE" {
		mod.clearEvents(w, r)
	} else {
		http.Error(w, "Bad Request", 400)
	}
}

func (mod *RestAPI) fileRoute(w http.ResponseWriter, r *http.Request) {
	mod.setSecurityHeaders(w)

	if !mod.checkAuth(r) {
		mod.setAuthFailed(w, r)
		return
	}

	fileName := r.URL.Query().Get("name")

	if fileName != "" && r.Method == "GET" {
		mod.readFile(fileName, w, r)
	} else if fileName != "" && r.Method == "POST" {
		mod.writeFile(fileName, w, r)
	} else {
		http.Error(w, "Bad Request", 400)
	}
}
