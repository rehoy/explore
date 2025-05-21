package main

import (
	"log"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gorilla/websocket"
	"github.com/rehoy/explore/balls"
)

func main() {
	// Connect to the websocket server
	wsURL := "ws://localhost:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	width, height := int32(800), int32(600)
	rl.InitWindow(width, height, "WebSocket Ball Client")
	defer rl.CloseWindow()

	var state []byte
	stateCh := make(chan []byte, 1)
	closeCh := make(chan struct{})

	// Goroutine to receive state from server
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				closeCh <- struct{}{}
				return
			}
			select {
			case stateCh <- msg:
			default:
			}
		}
	}()

	for !rl.WindowShouldClose() {
		select {
		case state = <-stateCh:
		case <-closeCh:
			os.Exit(0)
		default:
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		if len(state) > 0 {
			circles := balls.ImportState(state)
			for _, circle := range circles {
				r, g, b, a := circle.GetColor()
				rl.DrawCircle(int32(circle.X), int32(circle.Y), circle.Radius, rl.NewColor(r, g, b, a))
			}
		}

		rl.EndDrawing()
		time.Sleep(time.Second / 120) // Small sleep to avoid busy loop
	}

	// Optionally, send a message to close the simulation on exit
	// _ = conn.WriteMessage(websocket.TextMessage, []byte("close"))
}
