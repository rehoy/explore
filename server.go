package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
	"encoding/json"

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

	closeCh := make(chan struct{})

	// Listen for any message from client to close simulation
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				closeCh <- struct{}{}
				return
			}
			// On any message, signal to close
			closeCh <- struct{}{}
			return
		}
	}()

	count := 0
	for {
		select {
		case <-ticker.C:
			count++
			if count%600 == 0 {
				radius := float32(rand.Intn(50) + 10)
				x := uint16(rand.Intn(int(800)-int(2*radius)) + int(radius))
				y := uint16(rand.Intn(int(600)-int(2*radius)) + int(radius))

				velocity := balls.Velocity{
				X: 0.0,
				Y: 0.0,
				}

				context.AddCircle(x, y, radius, velocity)
				fmt.Println("Added circle:", x, y)
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
