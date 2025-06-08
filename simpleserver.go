package main

import (
	"github.com/rehoy/explore/server"
	"log"
	"net/http"
)

func main() {

	s := server.NewServer()
	go s.Logger.StartLogger()
	s.StartTestServer()
	http.HandleFunc("/ws", s.WsHandler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
