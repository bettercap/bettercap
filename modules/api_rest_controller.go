package modules

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bettercap/bettercap/session"
)

type CommandRequest struct {
	Command string `json:"cmd"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"msg"`
}

func setAuthFailed(w http.ResponseWriter) {
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

	events := session.I.Events.Sorted()
	nmax := len(events)
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

	toJSON(w, events[0:n])
}

func (api *RestAPI) clearEvents(w http.ResponseWriter, r *http.Request) {
	session.I.Events.Clear()
}

func (api *RestAPI) sessionRoute(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)

	if api.checkAuth(r) == false {
		setAuthFailed(w)
	} else if r.Method == "GET" {
		api.showSession(w, r)
	} else if r.Method == "POST" {
		api.runSessionCommand(w, r)
	} else {
		http.Error(w, "Bad Request", 400)
	}
}

func (api *RestAPI) eventsRoute(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)

	if api.checkAuth(r) == false {
		setAuthFailed(w)
	} else if r.Method == "GET" {
		api.showEvents(w, r)
	} else if r.Method == "DELETE" {
		api.clearEvents(w, r)
	} else {
		http.Error(w, "Bad Request", 400)
	}
}
