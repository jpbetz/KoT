package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"jpbetz/KoT/device-hub/service/types"
)

func websocketHandler(s *server, w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
	conn, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		log.Println(err)
		return
	}
	client := &WebsocketClientConn{websocketManager: s.websockets, conn: conn, send: make(chan []byte, 256)}
	client.websocketManager.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()

	// Send client full list of data to initialize with
	func () {
		s.mu.Lock()
		defer s.mu.Unlock()
		for _, device := range s.deviceHub.Devices {
			for _, input := range device.Inputs {
				m := &types.ValueChangedMessage{
					Path: device.ID + "." + input.ID,
					Value: input.Value,
				}
				data, err := json.Marshal(m)
				if err != nil {
					log.Printf("error: %v", err)
					continue
				}
				client.send <- data
			}
			for _, output := range device.Outputs {
				m := &types.ValueChangedMessage{
					Path: device.ID + "." + output.ID,
					Value: output.Value,
				}
				data, err := json.Marshal(m)
				if err != nil {
					log.Printf("error: %v", err)
					continue
				}
				client.send <- data
			}
		}
	}()
}


const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type WebsocketManager struct {
	// Registered clients.
	clients map[*WebsocketClientConn]bool

	// Inbound messages from the clients.
	broadcast chan *types.ValueChangedMessage

	// Register requests from the clients.
	register chan *WebsocketClientConn

	// Unregister requests from clients.
	unregister chan *WebsocketClientConn
}

func newWebsocketManager() *WebsocketManager {
	return &WebsocketManager{
		broadcast:  make(chan *types.ValueChangedMessage),
		register:   make(chan *WebsocketClientConn),
		unregister: make(chan *WebsocketClientConn),
		clients:    make(map[*WebsocketClientConn]bool),
	}
}

func (h *WebsocketManager) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			b, err := json.Marshal(message)
			if err != nil {
				log.Printf("error: %v", err)
				continue
			}
			for client := range h.clients {
				select {
				case client.send <- b:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebsocketClientConn struct {
	websocketManager *WebsocketManager
	conn *websocket.Conn
	send chan []byte
}

func (c *WebsocketClientConn) readPump() {
	defer func() {
		c.websocketManager.unregister <- c
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		msg := &types.ValueChangedMessage{}
		err = json.Unmarshal(message, msg)
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}
		// TODO: write the message to server and broadcast if we want to support websocket writes
	}
}

func (c *WebsocketClientConn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}