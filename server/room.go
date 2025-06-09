package server

import (
	"github.com/gorilla/websocket"
	"github.com/rehoy/explore/game"
	"log"
	"sync"
)

type Room struct {
	ID        string
	Game      game.Game
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex

	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	stop       chan struct{}
	roomsMu    sync.RWMutex
}

func NewRoom(id string, game game.Game) *Room {
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
			r.clientsMu.Unlock()
			log.Printf("Client joined room %s. Total clients: %d", r.ID, len(r.clients))
			r.Game.AddClient(conn)
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

func (r *Room) GetClientCount() int {
	r.clientsMu.Lock()
	defer r.clientsMu.Unlock()
	return len(r.clients)
}

func (r *Room) StopRoom() {
	if r.Game.IsRunning() {
		r.Game.Stop()
	}
}
