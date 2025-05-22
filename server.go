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

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	context := balls.MakeContext(800, 600)
	context.InitCircles(10)

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

	count := 0
	for {
		select {
		case <-ticker.C:
			count++
			if count%600 == 0 {
				var radius float32 = 15.0
				x := uint16(rand.Intn(int(800)-int(2*radius)) + int(radius))
				y := uint16(rand.Intn(int(600)-int(2*radius)) + int(radius))

				velocity := balls.Velocity{
					X: 0.0,
					Y: 0.0,
				}

				context.AddCircle(x, y, radius, velocity)
				fmt.Println("Added circle:", x, y)
			}
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
				fmt.Println("Added circle from client:", x, y)
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
	http.HandleFunc("/ws", wsHandler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
