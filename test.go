package main

import (
	"time"
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/rehoy/explore/balls"
)



func main() {
	context := balls.MakeContext(800, 600)
	context.InitCircles(10)
    rl.InitWindow(context.Width, context.Height, "Hello, raylib-go")
    defer rl.CloseWindow()

	stateCh := make(chan []byte, 1)
	go func () {
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()
		for range ticker.C {
			_ = context.UpdateCircles()
			state := context.ExportState()

			select {
			case stateCh <- state:
			default:
				fmt.Println("Missed a frame")
			}
		}
		
	}()
	
	var state []byte

    for !rl.WindowShouldClose() {
		select {
		case state = <-stateCh:
		default:
		}
        rl.BeginDrawing()
        rl.ClearBackground(rl.RayWhite)

		circles := balls.ImportState(state)
		for _, circle := range circles {
			r, g, b, a := circle.GetColor()
			rl.DrawCircle(int32(circle.X), int32(circle.Y), circle.Radius, rl.NewColor(r, g, b, a))

		}

        rl.EndDrawing()


    }
}