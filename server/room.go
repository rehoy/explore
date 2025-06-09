package server

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Room struct {
	ID        string
	Game      Game
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex

	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	stop       chan struct{}
	roomsMu    sync.RWMutex
}

func NewRoom(id string, game Game) *Room {
	r := &Room{
		ID:         id,
		Game:       game,
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		stop:       make(chan struct{}),
	}
	return r

}

func (r *Room) Run() {
	r.Game.Start()

	for {
		select {
		case conn := <-r.register:
			r.clientsMu.Lock()
			r.clients[conn] = true
			if err := r.Game.AddClient(conn); err != nil {
				delete(r.clients, conn)
				var e Event
				e.Type = "error"
				payloadBytes, _ := json.Marshal(err)
				e.Payload = payloadBytes
				_ = conn.WriteJSON(e)
				continue
			}
			log.Printf("Client joined room %s. Total clients: %d", r.ID, len(r.clients))
		case conn := <-r.unregister:
			r.clientsMu.Lock()
			if _, ok := r.clients[conn]; !ok {
				delete(r.clients, conn)
				r.Game.RemoveClient(conn)
				log.Printf("Client left room %s. Total clients: %d", r.ID, len(r.clients))
			}
			r.clientsMu.Unlock()
			if len(r.clients) == 0 {
				r.Game.Stop()
				close(r.stop)
				return
			}
		case <-r.stop:
			r.Game.Stop()
			log.Printf("Room %s stopped\n", r.ID)
			return

		}

	}
}

func (r *Room) getClientCount() int {
	r.clientsMu.Lock()
	defer r.clientsMu.Unlock()
	return len(r.clients)
}

func (r *Room) StopRoom() {
	if r.Game.IsRunning() {
		r.Game.Stop()
	}
}
