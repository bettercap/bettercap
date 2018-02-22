package modules

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bettercap/bettercap/session"
)

var (
	ApiUsername = ""
	ApiPassword = ""
)

type CommandRequest struct {
	Command string `json:"cmd"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"msg"`
}

func checkAuth(r *http.Request) bool {
	user, pass, _ := r.BasicAuth()
	// timing attack my ass
	if subtle.ConstantTimeCompare([]byte(user), []byte(ApiUsername)) != 1 {
		return false
	} else if subtle.ConstantTimeCompare([]byte(pass), []byte(ApiPassword)) != 1 {
		return false
	}
	return true
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

func showSession(w http.ResponseWriter, r *http.Request) {
	toJSON(w, session.I)
}

func runSessionCommand(w http.ResponseWriter, r *http.Request) {
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

func showEvents(w http.ResponseWriter, r *http.Request) {
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

func clearEvents(w http.ResponseWriter, r *http.Request) {
	session.I.Events.Clear()
}

func SessionRoute(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)

	if checkAuth(r) == false {
		setAuthFailed(w)
	} else if r.Method == "GET" {
		showSession(w, r)
	} else if r.Method == "POST" {
		runSessionCommand(w, r)
	} else {
		http.Error(w, "Bad Request", 400)
	}
}

func EventsRoute(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)

	if checkAuth(r) == false {
		setAuthFailed(w)
	} else if r.Method == "GET" {
		showEvents(w, r)
	} else if r.Method == "DELETE" {
		clearEvents(w, r)
	} else {
		http.Error(w, "Bad Request", 400)
	}
}
