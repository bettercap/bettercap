package modules

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/evilsocket/bettercap-ng/log"
)

func (api *RestAPI) setupRoutes() {
	http.HandleFunc("/api/session", api.sessRoute)
	http.HandleFunc("/api/events", api.eventsRoute)
}

func (api RestAPI) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	if api.Authenticated(w, r) == false {
		log.Warning("Unauthenticated access!")
		http.Error(w, "Not authorized", 401)
		return false
	}
	return true
}

func (api RestAPI) Authenticated(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(parts) != 2 {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}

	if pair[0] != api.username || pair[1] != api.password {
		return false
	}

	return true
}

func (api *RestAPI) sessRoute(w http.ResponseWriter, r *http.Request) {
	if api.checkAuth(w, r) == false {
		return
	}

	if r.Method == "GET" {
		js, err := json.Marshal(api.Session)
		if err != nil {
			log.Error("Error while returning session: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else if r.Method == "POST" && r.Body != nil {
		var req JSSessionRequest
		var res JSSessionResponse

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		err = api.Session.Run(req.Command)
		if err != nil {
			res.Error = err.Error()
		}
		js, err := json.Marshal(res)
		if err != nil {
			log.Error("Error while returning response: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else {
		http.Error(w, "Not Found", 404)
	}
}

func (api *RestAPI) eventsRoute(w http.ResponseWriter, r *http.Request) {
	if api.checkAuth(w, r) == false {
		return
	}

	if r.Method == "GET" {
		var err error

		events := api.Session.Events.Events()
		nmax := len(events)
		n := nmax

		keys, ok := r.URL.Query()["n"]
		if len(keys) == 1 && ok {
			sn := keys[0]
			n, err = strconv.Atoi(sn)
			if err == nil {
				if n > nmax {
					n = nmax
				}
			} else {
				n = nmax
			}
		}

		js, err := json.Marshal(events[0:n])
		if err != nil {
			log.Error("Error while returning events: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else if r.Method == "DELETE" {
		api.Session.Events.Clear()
		api.Session.Events.Add("sys.log.cleared", nil)
	} else {
		http.Error(w, "Not Found", 404)
	}
}
