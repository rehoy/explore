package main

import (
	"github.com/rehoy/explore/server"
	"log"
	"net/http"
	"time"
)

func main() {

	s := server.NewServer()
	go s.Logger.StartLogger()
	go func() {
		ticker := time.NewTicker(time.Second * 5)

		for {
			select {
			case <-ticker.C:
				s.PrintRoomInfo()
			}
		}

	}()
	http.HandleFunc("/ws", s.WsHandler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
