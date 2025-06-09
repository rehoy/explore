package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rehoy/explore/balls"
	"github.com/rehoy/explore/logger"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type MousePos struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type SetUsernamePayload struct {
	Name string `json:"name"`
}

type AddCirclePayload MousePos

type User struct {
	Name string `json:"name"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	rooms   map[string]*Room
	Logger  *logger.Logger
	roomsMu sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		rooms:  make(map[string]*Room),
		Logger: logger.NewLogger("server.log"),
	}
}

func (s *Server) sendEvent(conn *websocket.Conn, eventType string, payload interface{}) error {
	event := Event{
		Type: eventType,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.Logger.Log("could not marshal", eventType)
		return err
	}

	event.Payload = payloadBytes

	return conn.WriteJSON(event)
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

func (s *Server) handleEvent(event Event, context *balls.Context, conn *websocket.Conn) {

	switch event.Type {
	case "add_circle":
		s.Logger.Log("added circle")
		var p AddCirclePayload

		if err := json.Unmarshal(event.Payload, &p); err != nil {
			s.Logger.Log("Could not unmarshal payload")
		}

		radius := float32(rand.Intn(30) + 10)
		x := uint16(p.X)
		y := uint16(p.Y)
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

	case "set_userName":
		s.Logger.Log("Set username")
		var p SetUsernamePayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			log.Printf("Error unmarshaling set_username payload: %v", err)
			return
		}
		// Here you would associate the name `p.Name` with the connection.
		// For now, we'll just log it.
		log.Printf("User set name to: %s", p.Name)
		// You could send a confirmation back to the client
		s.sendEvent(conn, "username_accepted", p)
	default:
		s.Logger.Log("Event type", event, "not recognized")
	}
}

func (s *Server) GetorCreateRoom(roomID, gametype string) (*Room, error) {

	s.roomsMu.Lock()
	defer s.roomsMu.Unlock()

	room, ok := s.rooms[roomID]
	if ok {
		return room, nil
	}

	var game Game

	switch gametype {
	case "balls":
		ctx := balls.MakeContext(800, 600)
		game = NewBallsGameWrapper(roomID, &ctx)
	case "colors":
		s.Logger.Log("Not implemented colors yet")
		return nil, fmt.Errorf("Not implemeted yet")
	default:
		s.Logger.Log("not a recognized gametype", gametype)
		return nil, fmt.Errorf("Not recognized", gametype)
	}

	room = NewRoom(roomID, game)

	s.rooms[roomID] = room
	go room.Run()
	s.Logger.Log("Created and started new room:", roomID, "for game type:", gametype)
	return room, nil

}

func (s *Server) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// Get room from query, default to "default"
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		roomID = "default"
	}

	gameType := r.URL.Query().Get("game")
	if gameType == "" {
		gameType = "balls"
	}

	room, err := s.GetorCreateRoom(roomID, gameType)
	if err != nil {
		s.Logger.Log("Failed to create room", roomID, gameType)
		conn.Close()
		return
	}

	room.register <- conn
	defer func() {
		room.unregister <- conn
		conn.Close()
	}()

	go func() {

		for {

			_, msg, err := conn.ReadMessage()
			if err != nil {
				s.Logger.Log("Read error for client in room", roomID, err)
				break
			}

			if err := room.Game.HandleClientMessage(conn, msg); err != nil {
				s.Logger.Log("Error handling message in room", roomID, err)
			}
		}
	}()

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			state := room.Game.ExportState()
			if err := conn.WriteMessage(websocket.BinaryMessage, state); err != nil {
				s.Logger.Log("Write error to client in room %s: %v", roomID, err)
				return // Exit if cannot write to client
			}
		}

	}

}
