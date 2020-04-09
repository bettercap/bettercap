package api_rest

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

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

func (mod *RestAPI) streamEvent(ws *websocket.Conn, event session.Event) error {
	msg, err := json.Marshal(event)
	if err != nil {
		mod.Error("Error while creating websocket message: %s", err)
		return err
	}

	ws.SetWriteDeadline(time.Now().Add(writeWait))
	if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
		if !strings.Contains(err.Error(), "closed connection") {
			mod.Error("Error while writing websocket message: %s", err)
			return err
		}
	}

	return nil
}

func (mod *RestAPI) sendPing(ws *websocket.Conn) error {
	ws.SetWriteDeadline(time.Now().Add(writeWait))
	ws.SetReadDeadline(time.Now().Add(pongWait))
	if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
		mod.Error("Error while writing websocket ping message: %s", err)
		return err
	}
	return nil
}

func (mod *RestAPI) streamWriter(ws *websocket.Conn, w http.ResponseWriter, r *http.Request) {
	defer ws.Close()

	// first we stream what we already have
	events := session.I.Events.Sorted()
	n := len(events)
	if n > 0 {
		mod.Debug("Sending %d events.", n)
		for _, event := range events {
			if err := mod.streamEvent(ws, event); err != nil {
				return
			}
		}
	}

	session.I.Events.Clear()

	mod.Debug("Listening for events and streaming to ws endpoint ...")

	pingTicker := time.NewTicker(pingPeriod)
	listener := session.I.Events.Listen()
	defer session.I.Events.Unlisten(listener)

	for {
		select {
		case <-pingTicker.C:
			if err := mod.sendPing(ws); err != nil {
				return
			}
		case event := <-listener:
			if err := mod.streamEvent(ws, event); err != nil {
				return
			}
		case <-mod.quit:
			mod.Info("Stopping websocket events streamer ...")
			return
		}
	}
}

func (mod *RestAPI) streamReader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			mod.Warning("error reading message from websocket: %v", err)
			break
		}
	}
}

func (mod *RestAPI) startStreamingEvents(w http.ResponseWriter, r *http.Request) {
	ws, err := mod.upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			mod.Error("error while updating api.rest connection to websocket: %s", err)
		}
		return
	}

	mod.Debug("websocket streaming started for %s", r.RemoteAddr)

	go mod.streamWriter(ws, w, r)
	mod.streamReader(ws)
}
