package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gorilla/websocket"
	"github.com/rehoy/explore/balls"
)

type MousePos struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func main() {
	room := "one"
	game := "balls"
	if len(os.Args) > 1 {
		room = os.Args[1]
	}
	if len(os.Args) > 2 {
		game = os.Args[2]
	}
	wsURL := fmt.Sprintf("ws://localhost:8080/ws?room=%s&game=%s", room, game)
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
	spaceCh := make(chan struct{}, 1)

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

	go isPressed(spaceCh)

	for !rl.WindowShouldClose() {
		select {
		case <-spaceCh:
			mousePos := MousePos{
				X: float32(rl.GetMouseX()),
				Y: float32(rl.GetMouseY()),
			}
			fmt.Println("Adding a new circle at:", mousePos.X, mousePos.Y)
			event := Event{Type: "add_circle"}
			payloadBytes, err := json.Marshal(mousePos)
			if err != nil {
				log.Println("Could not marshal mouse pos", mousePos)
			}
			event.Payload = payloadBytes

			err = conn.WriteJSON(event)
			if err != nil {
				log.Println("Write error:", err)
				closeCh <- struct{}{}
				return
			}
		default:
			// no key press
		}

		select {
		case state = <-stateCh:
		case <-closeCh:
			os.Exit(0)
		default:
			// no new state
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

func isPressed(spaceCh chan struct{}) {
	for {
		if rl.IsKeyPressed(rl.KeySpace) {
			select {
			case spaceCh <- struct{}{}:
			default:
			}
			// Wait until key is released to avoid multiple triggers
			for rl.IsKeyDown(rl.KeySpace) {
				time.Sleep(time.Millisecond * 5)
			}
		}
		time.Sleep(time.Millisecond * 2)
	}
}
