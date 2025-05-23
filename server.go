package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rehoy/explore/balls"
)

type MousePos struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	rooms map[string]*balls.Context
}

func NewServer() *Server {
	return &Server{
		rooms: make(map[string]*balls.Context),
	}
}

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// Get room from query, default to "default"
	room := r.URL.Query().Get("room")
	if room == "" {
		room = "default"
	}

	// Get or create context for this room
	context, ok := s.rooms[room]
	if !ok {
		ctx := balls.MakeContext(800, 600)
		ctx.InitCircles(10)
		context = &ctx
		s.rooms[room] = context
	}

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	addCircleCh := make(chan MousePos, 8)
	closeCh := make(chan struct{})

	// Listen for messages from client to add circles
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				closeCh <- struct{}{}
				return
			}
			var mouse MousePos
			if err := json.Unmarshal(msg, &mouse); err == nil {
				addCircleCh <- mouse
			}
			// Ignore messages that can't be parsed as MousePos
		}
	}()

	for {
		select {
		case <-ticker.C:
			// Add circles from client mouse positions
			select {
			case mouse := <-addCircleCh:
				radius := float32(rand.Intn(30) + 10)
				x := uint16(mouse.X)
				y := uint16(mouse.Y)
				var vx, vy float32
				for vx == 0 {
					vx = float32(rand.Float64()*4 - 2)
				}
				for vy == 0 {
					vy = float32(rand.Float64()*4 - 2)
				}
				velocity := balls.Velocity{
					X: vx,
					Y: vy,
				}
				context.AddCircle(x, y, radius, velocity)
				fmt.Println("Added circle from client:", x, y, velocity)
			default:
				// no mouse event
			}
			_ = context.UpdateCircles()
			state := context.ExportState()
			err := conn.WriteMessage(websocket.BinaryMessage, state)
			if err != nil {
				log.Println("Write error:", err)
				return
			}
		case <-closeCh:
			log.Println("Received close signal from client")
			return
		}
	}
}

func main() {
	s := NewServer()
	http.HandleFunc("/ws", s.wsHandler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
