// Flippy is a proof of concept AI that interfaces with the Pao server
// but doesn't make any meaningful game decisions
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/arbrown/pao/game/command"
	"github.com/arbrown/pao/game/util"
	"github.com/gorilla/websocket"
)

type flippy struct {
	conn *websocket.Conn
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	fmt.Println("Hello, Flippy")
	host, port := os.Getenv("FLIPPYHOST"), os.Getenv("FLIPPYPORT")
	if port == "" {
		port = "2020"
	}
	bind := fmt.Sprintf("%v:%v", host, port)
	http.HandleFunc("/", flippyServe)
	http.ListenAndServe(bind, nil)
}

func flippyServe(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Request: %v\n", r)
	conn, err := upgrader.Upgrade(w, r, nil)
	fmt.Printf("Got a new customer!\n")
	if err != nil {
		fmt.Printf("FlippyErr = %v\n", err.Error())
		return
	}
	f := flippy{
		conn: conn,
	}
	go playFlippy(f)
	fmt.Printf("%v - Exiting Handler\n", time.Now())
}

func playFlippy(flippy flippy) {

	defer readLoop(flippy.conn)
	for {
		var com command.Command
		_, bytes, err := flippy.conn.ReadMessage()
		if err != nil {
			fmt.Printf("%v - Error reading from websocket: %v\n", time.Now(), err.Error())
			return
		}
		if err = json.Unmarshal(bytes, &com); err != nil {
			fmt.Printf("Error decoding websocket message: %v\n", bytes)
		}
		switch com.Action {
		case "board":
			var bc command.BoardCommand
			if err = json.Unmarshal(bytes, &bc); err != nil {
				fmt.Printf("Error decoding board command: %v\n", string(bytes))
			}
			flippy.processBoard(bc)
		default:
			fmt.Printf("Message: %v\n", string(bytes))
		}
	}
}

func (f *flippy) processBoard(bc command.BoardCommand) {
	gs := util.ParseGameState(bc)
	if !bc.YourTurn {
		return
	}
	if len(gs.RemainingPieces) == 0 {
		f.resign()
		return
	}
	for {
		rank, file := rand.Intn(len(gs.KnownBoard)), rand.Intn(len(gs.KnownBoard[0]))
		piece := gs.KnownBoard[rank][file]
		if piece == "?" {
			f.flip(rank, file)
			return
		}
	}
}

func (f *flippy) flip(rank, file int) {
	s := "?" + util.ToNotation(rank, file)
	com := command.Command{
		Action:   "move",
		Argument: s,
	}
	f.sendCommand(com)
}

func (f *flippy) resign() {
	com := command.Command{
		Action: "resign",
	}
	f.sendCommand(com)
}

func (f *flippy) sendCommand(c command.Command) {
	if err := f.conn.WriteJSON(c); err != nil {
		fmt.Printf("Error sending command: %v", err.Error())
	}
}

func readLoop(c *websocket.Conn) {
	fmt.Printf("%v - Closing conn\n", time.Now())
	for {
		if _, _, err := c.NextReader(); err != nil {
			c.Close()
			break
		}
	}
}
