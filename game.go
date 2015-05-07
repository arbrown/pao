package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type game struct {
	id                        string
	black, white              *player
	board, knownBoard         [][]string
	active                    bool
	commandChan               chan playerCommand
	currentPlayer, nextPlayer *player
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (g *game) join(w http.ResponseWriter, r *http.Request) bool {
	//fmt.Printf("w=%+v\nr=%+v\n", w, r)
	fmt.Printf("upgrader= %+v\n", upgrader)
	if g.currentPlayer == nil {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("Err = %v\n", err.Error())
		}
		g.currentPlayer = newPlayer(conn, g)
		fmt.Println("Joined as #1")
		go g.listenPlayer(g.currentPlayer)
		go g.startGame()
		return true
	} else if g.nextPlayer == nil {
		fmt.Println("Trying to join as #2")
		conn, err := upgrader.Upgrade(w, r, nil)
		fmt.Println("WS upgraded...")
		fmt.Printf("err=%v\n", err)
		if err != nil {
			fmt.Printf("Error joining as #2: %v\n", err.Error())
			return false
		}
		g.nextPlayer = newPlayer(conn, g)
		go g.listenPlayer(g.nextPlayer)
		return true
	}
	return false
}

func (g *game) startGame() {
	fmt.Println("game loop started")
	//board := makeNewGameBoard()
	g.knownBoard = generateUnknownBoard()
	for {
		select {
		case c := <-g.commandChan:
			// handle p1 move
			g.handleCommand(c)
		}
	}

}

func (g *game) handleCommand(c playerCommand) {
	fmt.Printf("Got command: %+v\n", c)
	if g.currentPlayer != c.p {
		r := command{Action: "info", Argument: "Not your turn!"}
		c.p.ws.WriteJSON(r)
		return
	}
	switch c.c.Action {
	case "chat":
		g.broadcastChat(c.p.name, c.c.Argument)
		g.currentPlayer, g.nextPlayer = g.nextPlayer, g.currentPlayer
	case "board?":
		g.broadcastBoard()
	}

}

func (g *game) broadcastBoard() {
	b, _ := json.Marshal(g.knownBoard)
	r := command{Action: "board", Argument: string(b)}
	g.broadcast(r)
}

func (g *game) broadcastChat(from, message string) {
	chat := chatMessage{Player: from, Message: message}
	b, _ := json.Marshal(chat)
	r := command{Action: "chat", Argument: string(b)}
	g.broadcast(r)
}

func (g *game) broadcast(v interface{}) {
	if g.currentPlayer != nil {
		g.currentPlayer.ws.WriteJSON(v)
	}
	if g.nextPlayer != nil {
		g.nextPlayer.ws.WriteJSON(v)
	}
}

func (g *game) listenPlayer(p *player) {
	//fmt.Printf("Listening to new player in game %s\n", g.id)
	for {
		var com command
		err := p.ws.ReadJSON(&com)
		//fmt.Println("got a message?")
		if err != nil {
			fmt.Printf("Error from player's messages: %v\n", err.Error())
			continue
		} else {
			b, jerr := json.Marshal(com)
			if jerr != nil {
				fmt.Println(jerr.Error())
			}
			fmt.Printf("Marshalled: %v\n", string(b))
			pc := playerCommand{c: com, p: p}
			g.commandChan <- pc
			// if wserr := p.ws.WriteJSON(com); wserr != nil {
			// 	fmt.Printf("Error sending JSON: %v\n", wserr.Error())
			// }
		}
	}
}

func newGame(id string) *game {
	return &game{
		id:          id,
		black:       nil,
		white:       nil,
		active:      false,
		commandChan: make(chan playerCommand),
	}
}

// func makeNewGameBoard() [][]rune {
//
// }
func generateUnknownBoard() [][]string {
	board := make([][]string, 4)
	for i := 0; i < 4; i++ {
		board[i] = make([]string, 8)
		for j := 0; j < 8; j++ {
			board[i][j] = "?"
		}
	}
	return board
}
