package game

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type DuelGame struct {
	numPlayers int
	ticker     *time.Ticker
	players    []int
	conns      map[int]*websocket.Conn
	RoomID     string
	stopChan   chan struct{}
	count      int
	started    bool
}

func NewDuelGame(roomID string) *DuelGame {
	duel := &DuelGame{
		numPlayers: 2,
		players:    make([]int, 0),
		conns:      make(map[int]*websocket.Conn),
		RoomID:     roomID,
		stopChan:   make(chan struct{}),
		count:      0,
		started:    false}

	return duel
}

func (d *DuelGame) Update() {
	threshold := 120
	if len(d.players) != 0 {
		if d.count == threshold-1 {
			for i := range d.players {
				d.players[i] = d.players[i] - 1
			}
		}
	}
	d.count = (d.count + 1) % threshold

	fmt.Println(d.players)

}

func (d *DuelGame) HandleClientMessage(conn *websocket.Conn, msg []byte) error {
	//When receiving message add a number to the corresponding
	return nil

}

func (d *DuelGame) ExportState() []byte { return make([]byte, 0) }

func (d *DuelGame) AddClient(conn *websocket.Conn) {
	if len(d.players) != d.numPlayers {
		d.conns[len(d.players)] = conn
		d.players = append(d.players, 20+len(d.players))
	}

}

func (d *DuelGame) RemoveClient(conn *websocket.Conn) {}

func (d *DuelGame) IsRunning() bool { return d.started }

func (d *DuelGame) Start() {

	d.ticker = time.NewTicker(time.Second / 60)
	d.started = true

	go func() {

		for {
			select {
			case <-d.ticker.C:
				d.Update()
			}
		}

	}()

}

func (d *DuelGame) Stop() {}

func (d *DuelGame) GetRoomID() string { return d.RoomID }
