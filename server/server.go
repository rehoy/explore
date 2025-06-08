package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rehoy/explore/balls"
	"github.com/rehoy/explore/logger"
)

type MousePos struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type User struct {
	Name string `json:"name"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	rooms     map[string]*balls.Context
	Logger    *logger.Logger
	connRooms map[*websocket.Conn]string // track which room each connection is in
}

func NewServer() *Server {
	return &Server{
		rooms:     make(map[string]*balls.Context),
		Logger:    logger.NewLogger("server.log"),
		connRooms: make(map[*websocket.Conn]string), // initialize map
	}
}

func (s *Server) startRoomSimulation(room string, ctx *balls.Context) {
	go func() {
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()
		for {
			<-ticker.C
			ctx.UpdateCircles()
		}
	}()
}

func (s *Server) WsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Track connection to room
	s.connRooms[conn] = room
	defer delete(s.connRooms, conn)

	// Get or create context for this room
	context, ok := s.rooms[room]
	if !ok {
		ctx := balls.MakeContext(800, 600)
		ctx.InitCircles(0) // Initialize with 10 circles
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
				s.Logger.Log(1, fmt.Sprintf("Added circle at (%f, %f) with radius %f", mouse.X, mouse.Y, radius))
			default:
				// no mouse event
			}

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

func (s *Server) StartTestServer() {

	rooms := []struct {
		name       string
		numCircles int
	}{
		{"one", 10},
		{"two", 5},
		{"three", 20},
	}

	for _, room := range rooms {
		if _, ok := s.rooms[room.name]; !ok {
			ctx := balls.MakeContext(800, 600)
			ctx.InitCircles(room.numCircles)
			s.rooms[room.name] = &ctx
			s.startRoomSimulation(room.name, &ctx)
			s.Logger.Log(0, "Starting simulation for room:", room.name)
		}
	}
}
