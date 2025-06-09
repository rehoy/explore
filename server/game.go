package server

import 

type Game interface {
	Update()
	HandleClientMessage(conn *websocket.Conn, msg []byte) error
	ExportState() []byte
	AddClient(conn *websocket.Conn)
	RemoveClient(conn *websocket.Conn)
	IsRunning() bool
	Start()
	Stop()
	GetRoomID() string
}
