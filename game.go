package main

import (
	"encoding/json"
	"fmt"
	"math"
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
	pieceToInt                map[string]int
	canAttack                 [][]bool
	gameOverChan              chan bool
	removeGameChan            chan *game
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (g *game) join(w http.ResponseWriter, r *http.Request, name string) bool {
	//fmt.Printf("w=%+v\nr=%+v\n", w, r)
	fmt.Printf("upgrader= %+v\n", upgrader)
	if g.currentPlayer == nil {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("Err = %v\n", err.Error())
		}
		g.currentPlayer = newPlayer(conn, g, name)
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
		g.nextPlayer = newPlayer(conn, g, name)
		go g.listenPlayer(g.nextPlayer)
		return true
	}
	return false
}

func (g *game) startGame() {
	defer func() {
		fmt.Println("Removing game from list")
		g.removeGameChan <- g
	}()
	fmt.Println("game loop started")
	defer func() { fmt.Println("game loop ended") }()
	rand.Seed(time.Now().UTC().UnixNano())
	g.knownBoard = generateUnknownBoard()
	for {
		select {
		case c := <-g.commandChan:
			// handle p1 move
			g.handleCommand(c)
			break
		case _ = <-g.gameOverChan:
			return
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
		color := "black"
		if c.p == g.red {
			color = "red"
		}
		g.broadcastChat(c.p.name, c.c.Argument, color)
	case "board?":
		g.broadcastBoard()
	case "move":
		if ok := g.tryMove(c); ok {
			// check for a victory
			if winner, won := g.checkVictory(); won {
				g.broadcastBoard()
				g.broadcastVictory(winner)
				g.endGame()
			}
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

func (g *game) broadcastChat(from, message, color string) {
	chat := chatCommand{Action: "chat", Player: from, Message: message, Color: color}
	//b, _ := json.Marshal(chat)
	//r := command{Action: "chat", Argument: string(b)}
	g.broadcast(chat)
}

func (g *game) broadcastVictory(victor *player) {
	fmt.Printf("I think the victor is: %+v\n", victor)
	if victor != nil && victor.ws != nil {
		fmt.Println("I told the victor he won")
		c := gameOverCommand{Action: "gameover", Message: "You win!", YouWin: true}
		victor.ws.WriteJSON(c)
	}
	lose := gameOverCommand{Action: "gameover", Message: "You lose!", YouWin: false}
	fmt.Printf("Victor: %+v, red: %+v, black: %+v\n", victor, g.red, g.black)
	if g.red == victor && g.black != nil {
		fmt.Println("I told black he lost")
		g.black.ws.WriteJSON(lose)
	}
	if g.black == victor && g.red != nil {
		fmt.Println("I told red he lost")
		g.red.ws.WriteJSON(lose)
	}
}

func (g *game) endGame() {
	// just close the web sockets and channels
	if g.red != nil {
		g.red.ws.Close()
	}
	if g.black != nil {
		g.black.ws.Close()
	}
	g.gameOverChan <- true
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
	fmt.Printf("Listening to new player {%+v} in game %s\n", p, g.id)
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
	fmt.Println("Stopping listen loop")
	p.ws.Close()
	g.endGame()
}

func newGame(id string, removeGameChan chan *game) *game {
	return &game{
		id:             id,
		black:          nil,
		red:            nil,
		active:         false,
		commandChan:    make(chan playerCommand),
		gameOverChan:   make(chan bool),
		removeGameChan: removeGameChan,
		remainingPieces: []string{
			"K", "k",
			"G", "G", "g", "g",
			"E", "E", "e", "e",
			"C", "C", "c", "c",
			"H", "H", "h", "h",
			"P", "P", "P", "P", "P", "Q", "Q",
			"p", "p", "p", "p", "p", "q", "q"},
		pieceToInt: map[string]int{
			"K": 6, "k": 6, "G": 5, "g": 5,
			"E": 4, "e": 4, "C": 3, "c": 3,
			"H": 2, "h": 2, "P": 1, "p": 1,
			"Q": 0, "q": 0,
		},
		canAttack: [][]bool{
			{true, true, true, true, true, true, true},      // Cannon
			{false, true, false, false, false, false, true}, // Pawn
			{true, true, true, false, false, false, false},  // Horse
			{true, true, true, true, false, false, false},   // Cart
			{true, true, true, true, true, false, false},    // Elephant
			{true, true, true, true, true, true, false},     // Guard
			{true, false, true, true, true, true, true},     // King
		},
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
	if !g.currentPlayerOwns(srcRank, srcFile) || g.currentPlayerOwns(tgtRank, tgtFile) {
		return false
	}

	srcPiece, tgtPiece := g.knownBoard[srcRank][srcFile], g.knownBoard[tgtRank][tgtFile]
	if tgtPiece != "." {
		// Not an empty space, need to check if we can attack it
		if tgtPiece == "?" {
			// trying to attack an unflipped piece?  You monster!
			return false
		}
		srcPower, tgtPower := g.pieceToInt[srcPiece], g.pieceToInt[tgtPiece]
		if !g.canAttack[srcPower][tgtPower] {
			return false
		}
	}

	// validate the legality of the move
	if (srcPiece != "Q" && srcPiece != "q") || tgtPiece == "." {
		// no cannon attack involved, move must be directly adjacent
		if (math.Abs(float64(srcRank-tgtRank)) + math.Abs(float64(srcFile-tgtFile))) != 1 {
			// invalid move
			return false
		}
	} else {
		// now the harder part
		moveRank, moveFile := srcRank-tgtRank, srcFile-tgtFile
		// One of those should be 0
		if moveRank != 0 && moveFile != 0 {
			// diagonal cannon shot... very sneaky
			return false
		}
		// walk from one end to the other, expecting one piece exactly
		hopped := 0
		if moveRank != 0 {
			for i := math.Min(float64(srcRank), float64(tgtRank)) + 1; i < math.Max(float64(srcRank), float64(tgtRank)); i++ {
				if g.knownBoard[int(i)][srcFile] != "." {
					hopped++
				}
			} // Can SOMEONE explain to me the need for all the casting???
		} else {
			for i := math.Min(float64(srcFile), float64(tgtFile)) + 1; i < math.Max(float64(srcFile), float64(tgtFile)); i++ {
				if g.knownBoard[srcRank][int(i)] != "." {
					hopped++
				}
			}
		}
		if hopped != 1 {
			return false
		}
	}

	// at this point you should be able to make a move... I hope
	g.knownBoard[srcRank][srcFile], g.knownBoard[tgtRank][tgtFile] = ".", srcPiece

	return true
}

func (g *game) checkVictory() (victor *player, won bool) {
	defer func() {
		fmt.Printf("Game over =%v because:\n", won)
		fmt.Printf("remainingPieces: %v\n", g.remainingPieces)
	}()
	// is black out of pieces?
	blackRemains, redRemains := false, false
	for i := 0; i < 4 && !(blackRemains && redRemains); i++ {
		for j := 0; j < 8 && !(blackRemains && redRemains); j++ {
			p := g.knownBoard[i][j]
			redRemains = redRemains || isRed(p)
			blackRemains = blackRemains || isBlack(p)
			//fmt.Printf("Examined {%v}.  Red: %v Black %v\n", p, isRed(p), isBlack(p))
		}
	}

	for _, p := range g.remainingPieces {
		if redRemains && blackRemains {
			break
		}
		redRemains = redRemains || isRed(p)
		blackRemains = blackRemains || isBlack(p)
	}

	if redRemains && !blackRemains {
		victor = g.red
		won = true
		return
	}
	if blackRemains && !redRemains {
		victor = g.black
		won = true
		return
	}
	fmt.Printf("redRemains: %v, blackRemains:%v\n", redRemains, blackRemains)
	victor = nil
	won = false
	return
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

func isRed(piece string) bool {
	switch piece {
	case "k", "g", "e", "c", "h", "p", "q":
		return true
	default:
		return false
	}
}
func isBlack(piece string) bool {
	switch piece {
	case "K", "G", "E", "C", "H", "P", "Q":
		return true
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
