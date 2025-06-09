package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rehoy/explore/balls" // Your existing balls package
	"sync"
)

// BallsGameWrapper implements the Game interface for the balls simulation.
type BallsGameWrapper struct {
	roomID    string
	ctx       *balls.Context // The actual balls simulation context
	ticker    *time.Ticker
	stopCh    chan struct{}
	isRunning bool
	// In a real pub/sub, you'd have a channel for broadcasted state.
	// For now, it will just update its own context.
	clientsMux sync.RWMutex
	clients    map[*websocket.Conn]bool // Keep track of clients for direct sends if needed
}

func NewBallsGameWrapper(roomID string, ctx *balls.Context) *BallsGameWrapper {
	return &BallsGameWrapper{
		roomID:  roomID,
		ctx:     ctx,
		stopCh:  make(chan struct{}),
		clients: make(map[*websocket.Conn]bool), // Initialize map
	}
}

// --- Implementation of the Game interface ---

func (b *BallsGameWrapper) Update() {
	b.ctx.UpdateCircles() // This should be race-condition safe due to mutex in balls.Context
}

func (b *BallsGameWrapper) HandleClientMessage(conn *websocket.Conn, msg []byte) error {
	var event Event // Assuming the event envelope
	if err := json.Unmarshal(msg, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	switch event.Type {
	case "add_circle":
		var p AddCirclePayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return fmt.Errorf("error unmarshaling add_circle payload: %w", err)
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
		velocity := balls.Velocity{X: vx, Y: vy}
		b.ctx.AddCircle(x, y, radius, velocity)
		log.Printf("[%s] Added circle at (%f, %f)", b.roomID, p.X, p.Y)
	default:
		return fmt.Errorf("unsupported event type for BallsGame: %s", event.Type)
	}
	return nil
}

func (b *BallsGameWrapper) ExportState() []byte {
	return b.ctx.ExportState() // This should be race-condition safe
}

func (b *BallsGameWrapper) AddClient(conn *websocket.Conn) {
	b.clientsMux.Lock()
	defer b.clientsMux.Unlock()
	b.clients[conn] = true
	// Logic for initializing client (e.g., sending current state)
}

func (b *BallsGameWrapper) RemoveClient(conn *websocket.Conn) {
	b.clientsMux.Lock()
	defer b.clientsMux.Unlock()
	delete(b.clients, conn)
	// Logic for cleanup (e.g., if a player leaves in a turn-based game)
}

func (b *BallsGameWrapper) IsRunning() bool {
	return b.isRunning
}

func (b *BallsGameWrapper) Start() {
	if b.isRunning {
		return
	}
	b.ticker = time.NewTicker(time.Second / 60)
	b.stopCh = make(chan struct{}) // Re-initialize for fresh start
	b.isRunning = true
	go func() {
		defer b.ticker.Stop()
		log.Printf("[%s] Balls game simulation started.", b.roomID)
		for {
			select {
			case <-b.ticker.C:
				b.Update()
				// In a pub/sub model, the game would broadcast state here:
				// b.broadcastState(b.ExportState())
			case <-b.stopCh:
				log.Printf("[%s] Balls game simulation stopped.", b.roomID)
				b.isRunning = false
				return
			}
		}
	}()
}

func (b *BallsGameWrapper) Stop() {
	if !b.isRunning {
		return
	}
	close(b.stopCh)
}

func (b *BallsGameWrapper) GetRoomID() string {
	return b.roomID
}

// --- End of Game interface implementation ---

/*
// Example for future pub/sub broadcasting logic within the game:
func (b *BallsGameWrapper) broadcastState(state []byte) {
    b.clientsMux.RLock()
    defer b.clientsMux.RUnlock()
    for conn := range b.clients {
        // This should be done carefully to not block the game loop.
        // Might need a non-blocking send or send on a dedicated client channel.
        err := conn.WriteMessage(websocket.BinaryMessage, state)
        if err != nil {
            log.Printf("Error broadcasting state to client: %v", err)
            // Handle error, maybe remove client if connection is broken
        }
    }
}
*/
