package modules

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"

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

func (api *RestAPI) startStreamingEvents(w http.ResponseWriter, r *http.Request) {
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
}
