package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/jpbetz/KoT/simulator/service/types"
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

	go client.writeSender()

	// Send client full list of data to initialize with
	func () {
		s.mu.Lock()
		defer s.mu.Unlock()
		for _, device := range s.devices {
			for _, input := range device.Status.ObservedInputs {
				if moduleName, ok := s.deviceModules[device.Name]; ok {
					m := &types.EventMessage{
						Type: "value",
						Path:  moduleName + "." + device.Name + "." + input.Name,
						Value: input.Value.AsDec().String(),
					}
					data, err := json.Marshal(m)
					if err != nil {
						log.Printf("error: %v", err)
						continue
					}
					client.send <- data
				}
			}
			for _, output := range device.Status.Outputs {
				if moduleName, ok := s.deviceModules[device.Name]; ok {
					m := &types.EventMessage{
						Type: "value",
						Path:  moduleName + "." + device.Name + "." + output.Name,
						Value: output.Value.AsDec().String(),
					}
					data, err := json.Marshal(m)
					if err != nil {
						log.Printf("error: %v", err)
						continue
					}
					client.send <- data
				}
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
)

type WebsocketManager struct {
	// Registered clients.
	clients map[*WebsocketClientConn]bool

	// Inbound messages from the clients.
	broadcast chan *types.EventMessage

	// Register requests from the clients.
	register chan *WebsocketClientConn

	// Unregister requests from clients.
	unregister chan *WebsocketClientConn
}

func newWebsocketManager() *WebsocketManager {
	return &WebsocketManager{
		broadcast:  make(chan *types.EventMessage),
		register:   make(chan *WebsocketClientConn),
		unregister: make(chan *WebsocketClientConn),
		clients:    make(map[*WebsocketClientConn]bool),
	}
}

func (h *WebsocketManager) SendValueChanged(path string, quantity resource.Quantity) {
	msg := &types.EventMessage{
		Type: "value",
		Path:  path,
		Value: quantity.AsDec().String(),
	}
	h.broadcast <- msg
}

func (h *WebsocketManager) SendModuleCreated(moduleName string) {
	msg := &types.EventMessage{
		Type: "module-created",
		Path:  moduleName,
	}
	h.broadcast <- msg
}

func (h *WebsocketManager) SendModuleDeleted(moduleName string) {
	msg := &types.EventMessage{
		Type: "module-deleted",
		Path:  moduleName,
	}
	h.broadcast <- msg
}

func (h *WebsocketManager) SendModuleUpdated(moduleName string) {
	msg := &types.EventMessage{
		Type: "module-updated",
		Path:  moduleName,
	}
	h.broadcast <- msg
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

func (c *WebsocketClientConn) writeSender() {
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
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Printf("error writing close message to websocket: %v", err)
				}
				return
			}

			func() {
				w, err := c.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				defer func() {
					if err := w.Close(); err != nil {
						return
					}
				}()
				_, err = w.Write(message)
				if err != nil {
					log.Printf("error writing message to websocket: %v", err)
					return
				}
				n := len(c.send)
				for i := 0; i < n; i++ {
					_, err := w.Write([]byte{'\n'})
					if err != nil {
						log.Printf("error writing message to websocket: %v", err)
					}
					_, err = w.Write(<-c.send)
					if err != nil {
						log.Printf("error writing message to websocket: %v", err)
					}
				}
			}()
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}