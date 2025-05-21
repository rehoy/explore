package main

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/rehoy/explore/balls"
)



func main() {
	context := balls.MakeContext(800, 600)
	context.InitCircles(10)
    rl.InitWindow(context.Width, context.Height, "Hello, raylib-go")
    defer rl.CloseWindow()

	stateCh := make(chan []balls.Circle, 1)
	go func () {
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()
		for range ticker.C {
			circles := context.UpdateCircles()
			select {
			case stateCh <- circles:
			default:
			}
		}
		
	}()
	
	var circles []balls.Circle

    for !rl.WindowShouldClose() {
		select {
		case circles = <-stateCh:
		default:
		}
        rl.BeginDrawing()
        rl.ClearBackground(rl.RayWhite)

		
		for _, circle := range circles {
			r, g, b, a := circle.GetColor()
			rl.DrawCircle(int32(circle.X), int32(circle.Y), circle.Radius, rl.NewColor(r, g, b, a))

		}

        rl.EndDrawing()


    }
}