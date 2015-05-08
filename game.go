package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type game struct {
	id                        string
	black, red                *player
	knownBoard                [][]string
	remainingPieces           []string
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
	rand.Seed(time.Now().UTC().UnixNano())
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
	// if g.currentPlayer != c.p {
	// 	r := command{Action: "info", Argument: "Not your turn!"}
	// 	c.p.ws.WriteJSON(r)
	// 	return
	// }
	switch c.c.Action {
	case "chat":
		g.broadcastChat(c.p.name, c.c.Argument)
	case "board?":
		g.broadcastBoard()
	case "move":
		if ok := g.tryMove(c); ok {
			// move successful, swap players
			g.currentPlayer, g.nextPlayer = g.nextPlayer, g.currentPlayer
			g.broadcastBoard()
		}
	}
}

func (g *game) broadcastBoard() {
	r := boardCommand{Action: "board", Board: g.knownBoard, YourTurn: g.nextPlayer != nil}
	if g.currentPlayer != nil {
		g.currentPlayer.ws.WriteJSON(r)
	}
	if g.nextPlayer != nil {
		r.YourTurn = false
		g.nextPlayer.ws.WriteJSON(r)
	}
}

func (g *game) broadcastChat(from, message string) {
	chat := chatCommand{Action: "chat", Player: from, Message: message}
	//b, _ := json.Marshal(chat)
	//r := command{Action: "chat", Argument: string(b)}
	g.broadcast(chat)
}

func (g *game) broadcast(v interface{}) {
	if g.currentPlayer != nil {
		g.currentPlayer.ws.WriteJSON(v)
	}
	if g.nextPlayer != nil {
		g.nextPlayer.ws.WriteJSON(v)
	}
}

func (g *game) broadcastColors() {
	color := colorCommand{Action: "color", Color: "red"}
	fmt.Printf("Want to broadcast colors to:\n%+v\n%+v\n", g.red, g.black)
	if g.red != nil && g.red.ws != nil {
		g.red.ws.WriteJSON(color)
	}
	color.Color = "black"
	if g.black != nil && g.black.ws != nil {
		g.black.ws.WriteJSON(color)
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
			break
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
	p.ws.Close()
	g.quit()
}

func (g *game) quit() {
	close(g.commandChan)
}

func newGame(id string) *game {
	return &game{
		id:          id,
		black:       nil,
		red:         nil,
		active:      false,
		commandChan: make(chan playerCommand),
		remainingPieces: []string{
			"K", "k",
			"G", "G", "g", "g",
			"E", "E", "e", "e",
			"C", "C", "c", "c",
			"H", "H", "h", "h",
			"P", "P", "P", "P", "P", "Q", "Q",
			"p", "p", "p", "p", "p", "q", "q"},
	}
}

func (g *game) tryMove(pc playerCommand) bool {
	if pc.p != g.currentPlayer {
		return false
	}
	move, err := parseMove(pc.c.Argument)
	if err != nil {
		fmt.Printf("Couldn't parse move: %v\n", err.Error())
		return false
	}
	if !(move.isFlip || move.isValid()) {
		return false
	}

	// now actually do the move
	if move.isFlip {
		if ok := g.flip(move); !ok {
			return false
		}
	} else if ok := g.performMove(move); !ok {
		return false
	}
	return true
}

func (g *game) flip(m *move) bool {
	srcFile, srcRank, _, _ := m.getCoords()
	if g.knownBoard[srcRank][srcFile] != "?" {
		//can't flip that!
		return false
	}
	// get a random piece from the remaining pieces
	index := rand.Intn(len(g.remainingPieces))
	piece := g.remainingPieces[index]
	g.knownBoard[srcRank][srcFile] = piece
	//fmt.Printf("Flipped one of %d remaining pieces!\n", len(g.remainingPieces))
	//fmt.Printf("Remaining Pieces:\n%v\n", g.remainingPieces)
	if len(g.remainingPieces) == 32 {
		//First move! Assign colors based on piece
		switch piece {
		case "K", "G", "E", "C", "H", "P", "Q":
			g.black = g.currentPlayer
			g.red = g.nextPlayer
			break
		case "k", "g", "e", "c", "h", "p", "q":
			g.red = g.currentPlayer
			g.black = g.nextPlayer
		}
		g.broadcastColors()
	}
	g.remainingPieces = append(g.remainingPieces[:index], g.remainingPieces[index+1:]...)
	return true
}

func (g *game) performMove(m *move) bool {
	srcFile, srcRank, tgtFile, tgtRank := m.getCoords()
	if !g.currentPlayerOwns(srcRank, srcFile) {
		return false
	}
	if tgtFile == tgtRank {
		return false
	}
	srcPiece, _ := g.knownBoard[srcRank][srcFile], g.knownBoard[tgtRank][tgtFile]

	// for now, just perform the move without checking to see if it is fair
	g.knownBoard[srcRank][srcFile], g.knownBoard[tgtRank][tgtFile] = ".", srcPiece

	return true
}

func (g *game) currentPlayerOwns(rank, file int) bool {
	switch g.knownBoard[rank][file] {
	case "K", "G", "E", "C", "H", "P", "Q":
		return g.currentPlayer == g.black
	case "k", "g", "e", "c", "h", "p", "q":
		return g.currentPlayer == g.red
	default:
		return false
	}
}

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
