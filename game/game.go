package game

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/apexskier/httpauth"
	"github.com/arbrown/pao/game/gamestate"
	"github.com/gorilla/websocket"
)

// Game is a struct that represents the state and connections of a multiplayer
// ban qi game that the server is hosting.
type Game struct {
	ID                        string
	gameState                 gamestate.Gamestate
	black, red                *player
	lastMove                  []string
	lastDead                  string
	active                    bool
	commandChan               chan playerCommand
	db                        *sql.DB
	CurrentPlayer, NextPlayer *player
	pieceToInt                map[string]int
	canAttack                 [][]bool
	gameOverChan              chan bool
	removeGameChan            chan *Game
	kibitzers                 []*player
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Join causes a connection to join a game as a websocket and player
func (g *Game) Join(w http.ResponseWriter, r *http.Request, name string, user *httpauth.UserData) bool {
	//fmt.Printf("w=%+v\nr=%+v\n", w, r)
	fmt.Printf("upgrader= %+v\n", upgrader)
	if g.CurrentPlayer == nil {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("Err = %v\n", err.Error())
			return false
		}
		g.CurrentPlayer = newPlayer(conn, g, name, user, false)
		fmt.Println("Joined as #1")
		go g.listenPlayer(g.CurrentPlayer)
		go g.startGame()
		return true
	} else if g.NextPlayer == nil {
		fmt.Println("Trying to join as #2")
		conn, err := upgrader.Upgrade(w, r, nil)
		fmt.Println("WS upgraded...")
		fmt.Printf("err=%v\n", err)
		if err != nil {
			fmt.Printf("Error joining as #2: %v\n", err.Error())
			return false
		}
		g.NextPlayer = newPlayer(conn, g, name, user, false)
		go g.listenPlayer(g.NextPlayer)
		return true
	} else {
		g.JoinKibitz(w, r, name, user)
	}
	return false
}

func (g *Game) JoinKibitz(w http.ResponseWriter, r *http.Request, name string, user *httpauth.UserData) bool {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Err = %v\n", err.Error())
		return false
	}
	kibitzer := newPlayer(conn, g, name, user, true)
	g.kibitzers = append(g.kibitzers, kibitzer)
	go g.listenPlayer(kibitzer)
	return true
}

func (g *Game) removeKibitzer(p *player) {
	var newKibitzers []*player
	for i, k := range g.kibitzers {
		if k == p {
			newKibitzers = append(g.kibitzers[:i], g.kibitzers[i+1:]...)
			break
		}
	}
	g.kibitzers = newKibitzers
}

func (g *Game) closeWebSockets() {
	if g.red != nil {
		g.red.ws.Close()
	}
	if g.black != nil {
		g.black.ws.Close()
	}
	for _, k := range g.kibitzers {
		go readLoop(k.ws)
	}
}

func readLoop(c *websocket.Conn) {
	for {
		if _, _, err := c.NextReader(); err != nil {
			c.Close()
			break
		}
	}
}

func (g *Game) startGame() {
	defer g.endGame()

	fmt.Println("game loop started")
	defer func() { fmt.Println("game loop ended") }()
	rand.Seed(time.Now().UTC().UnixNano())
	g.gameState.KnownBoard = generateUnknownBoard()
	for {
		fmt.Println("Listening!")
		select {
		case c := <-g.commandChan:
			fmt.Println("Heard a command")
			// handle p1 move
			g.handleCommand(c)
			fmt.Println("Done handling command")
			break
		case _ = <-g.gameOverChan:
			fmt.Println("Got a message on gameOverChan")
			return
		}
	}

}

func (g *Game) handleCommand(c playerCommand) {
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
		if c.p.kibitzer == true {
			color = "teal"
		}
		g.broadcastChat(c.p, c.c.Argument, color)
	case "board?":
		g.broadcastBoard()
	case "move":
		if ok := g.tryMove(c); ok {
			// check for a victory
			if winner, won := g.checkVictory(); won {
				g.broadcastBoard()
				g.broadcastVictory(winner)
				fmt.Println("Reporting end of game from handleCommand - move")
				g.endGame()
				fmt.Println("Trying to return from handle command")
				return
			}
			// move successful, swap players
			g.CurrentPlayer, g.NextPlayer = g.NextPlayer, g.CurrentPlayer
			g.broadcastBoard()
		}
	case "resign":
		g.resign(c.p)
	}
}

func (g *Game) resign(p *player) {
	if p.kibitzer {
		g.suggestResign(p)
		return
	}
	if g.red == nil {
		return
	}
	if p == g.CurrentPlayer {
		g.broadcastVictory(g.NextPlayer)
	} else {
		g.broadcastVictory(g.CurrentPlayer)
	}

	g.endGame()
}

func (g *Game) suggestResign(p *player) {
	g.broadcastChat(p, g.getTaunt(), "darkcyan")
}

func (g *Game) getTaunt() string {
	return taunts[rand.Intn(len(taunts))]
}

func (g *Game) broadcastBoard() {
	numPlayers := 0
	if g.CurrentPlayer != nil {
		numPlayers += 1
	}
	if g.NextPlayer != nil {
		numPlayers += 1
	}
	r := boardCommand{
		Action:     "board",
		Board:      g.gameState.KnownBoard,
		YourTurn:   g.NextPlayer != nil,
		Dead:       g.gameState.DeadPieces,
		LastDead:   g.lastDead,
		LastMove:   g.lastMove,
		WhoseTurn:  g.CurrentPlayer.Name,
		TurnColor:  "green",
		NumPlayers: numPlayers}

	if g.CurrentPlayer == g.red {
		r.TurnColor = "red"
	}
	if g.CurrentPlayer == g.black {
		r.TurnColor = "black"
	}

	if g.CurrentPlayer != nil {
		g.CurrentPlayer.ws.WriteJSON(r)
	}
	if g.NextPlayer != nil {
		r.YourTurn = false
		g.NextPlayer.ws.WriteJSON(r)
	}
	for _, k := range g.kibitzers {
		k.ws.WriteJSON(r)
	}
}

func (g *Game) broadcastChat(from *player, message, color string) {
	fmt.Printf("Chat from: %+v\n", from)
	chat := chatCommand{Action: "chat", Player: from.Name, Message: message, Color: color, Auth: from.user != nil && from.user.Username == from.Name}
	//b, _ := json.Marshal(chat)
	//r := command{Action: "chat", Argument: string(b)}
	g.broadcast(chat)
}

func (g *Game) broadcastVictory(victor *player) {
	fmt.Printf("I think the victor is: %+v\n", victor)
	var loser *player
	var winColor string
	if g.black == victor {
		loser = g.red
		winColor = "black"
	} else {
		loser = g.black
		winColor = "red"
	}
	if victor != nil && victor.ws != nil {
		fmt.Println("I told the victor he won")
		c := gameOverCommand{Action: "gameover", Message: "You win!", YouWin: true}
		victor.ws.WriteJSON(c)
	}
	lose := gameOverCommand{Action: "gameover", Message: "You lose!", YouWin: false}
	fmt.Printf("Victor: %+v, red: %+v, black: %+v\n", victor, g.red, g.black)
	if g.red == victor && g.black != nil {
		loser = g.black
		fmt.Println("I told black he lost")
		g.black.ws.WriteJSON(lose)
	}
	if g.black == victor && g.red != nil {
		fmt.Println("I told red he lost")
		g.red.ws.WriteJSON(lose)
	}
	c := gameOverCommand{Action: "gameover", Message: "Game Over!", YouWin: false}
	for _, k := range g.kibitzers {
		k.ws.WriteJSON(c)
	}
	g.reportVictory(victor, loser, winColor)
}

func (g *Game) reportVictory(victor, loser *player, winColor string) {
	var (
		victorName, loserName string
	)
	if victor.user != nil {
		victorName = victor.user.Username
	}
	if loser.user != nil {
		loserName = loser.user.Username
	}

	if g.db == nil {
		return
	}
	_, err := g.db.Exec(`insert into completedGames(winner, loser, winColor) values ($1,$2,$3)`, victorName, loserName, winColor)
	if err != nil {
		fmt.Printf("Could not insert row into db: %s\n", err)
	}
}

func (g *Game) endGame() {
	g.removeGameChan <- g
	fmt.Printf("gameOverChan: %v\n", g.gameOverChan)
	g.gameOverChan <- true
	g.closeWebSockets()
	fmt.Println("Trying to return from endGame")
	return
}

func (g *Game) broadcast(v interface{}) {
	if g.CurrentPlayer != nil {
		g.CurrentPlayer.ws.WriteJSON(v)
	}
	if g.NextPlayer != nil {
		g.NextPlayer.ws.WriteJSON(v)
	}
	for _, k := range g.kibitzers {
		k.ws.WriteJSON(v)
	}
}

func (g *Game) broadcastColors() {
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

func (g *Game) listenPlayer(p *player) {
	fmt.Printf("Listening to new player {%+v} in game %s\n", p, g.ID)
	for {
		var com command
		err := p.ws.ReadJSON(&com)
		fmt.Println("got a message?")
		if err != nil {
			fmt.Printf("Error from player's messages: %v\n", err.Error())
			if p != g.red && p != g.black {
				g.removeKibitzer(p)
			}
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
	go readLoop(p.ws)
	if p == g.CurrentPlayer || p == g.NextPlayer {
		g.endGame()
	}
}

//NewGame returns a newly initialized game struct
func NewGame(id string, removeGameChan chan *Game, db *sql.DB) *Game {
	return &Game{
		ID:           id,
		black:        nil,
		red:          nil,
		active:       false,
		commandChan:  make(chan playerCommand),
		gameOverChan: make(chan bool, 3), // I think this masks a bug
		// but I'm not sure how to fix it atm...  I need a go expert.
		removeGameChan: removeGameChan,
		db:             db,
		gameState: gamestate.Gamestate{
			RemainingPieces: []string{
				"K", "k",
				"G", "G", "g", "g",
				"E", "E", "e", "e",
				"C", "C", "c", "c",
				"H", "H", "h", "h",
				"P", "P", "P", "P", "P", "Q", "Q",
				"p", "p", "p", "p", "p", "q", "q"},
			DeadPieces: make([]string, 0),
		},
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

func (g *Game) tryMove(pc playerCommand) bool {
	if pc.p != g.CurrentPlayer {
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
		g.lastMove = nil
		g.lastMove = append(g.lastMove, move.source)
	} else {
		ok, deadPiece := g.performMove(move)
		if !ok {
			return false
		}
		g.lastMove = nil
		g.lastMove = append(g.lastMove, move.source)
		if deadPiece != "" {
			g.lastDead = deadPiece
		}
	}
	if move.target != "" {
		g.lastMove = append(g.lastMove, move.target)
	}

	return true
}

func (g *Game) flip(m *move) bool {
	srcFile, srcRank, _, _ := m.getCoords()
	if g.gameState.KnownBoard[srcRank][srcFile] != "?" {
		//can't flip that!
		return false
	}
	// get a random piece from the remaining pieces
	index := rand.Intn(len(g.gameState.RemainingPieces))
	piece := g.gameState.RemainingPieces[index]
	g.gameState.KnownBoard[srcRank][srcFile] = piece
	if len(g.gameState.RemainingPieces) == 32 {
		//First move! Assign colors based on piece
		switch piece {
		case "K", "G", "E", "C", "H", "P", "Q":
			g.black = g.CurrentPlayer
			g.red = g.NextPlayer
			break
		case "k", "g", "e", "c", "h", "p", "q":
			g.red = g.CurrentPlayer
			g.black = g.NextPlayer
		}
		g.broadcastColors()
	}
	g.gameState.RemainingPieces = append(g.gameState.RemainingPieces[:index], g.gameState.RemainingPieces[index+1:]...)
	return true
}

func (g *Game) performMove(m *move) (bool, string) {
	srcFile, srcRank, tgtFile, tgtRank := m.getCoords()
	if !g.currentPlayerOwns(srcRank, srcFile) || g.currentPlayerOwns(tgtRank, tgtFile) {
		return false, ""
	}

	srcPiece, tgtPiece := g.gameState.KnownBoard[srcRank][srcFile], g.gameState.KnownBoard[tgtRank][tgtFile]
	if tgtPiece != "." {
		// Not an empty space, need to check if we can attack it
		if tgtPiece == "?" {
			// trying to attack an unflipped piece?  You monster!
			return false, ""
		}
		srcPower, tgtPower := g.pieceToInt[srcPiece], g.pieceToInt[tgtPiece]
		if !g.canAttack[srcPower][tgtPower] {
			return false, ""
		}
	}

	// validate the legality of the move
	if (srcPiece != "Q" && srcPiece != "q") || tgtPiece == "." {
		// no cannon attack involved, move must be directly adjacent
		if (math.Abs(float64(srcRank-tgtRank)) + math.Abs(float64(srcFile-tgtFile))) != 1 {
			// invalid move
			return false, ""
		}
	} else {
		// now the harder part
		moveRank, moveFile := srcRank-tgtRank, srcFile-tgtFile
		// One of those should be 0
		if moveRank != 0 && moveFile != 0 {
			// diagonal cannon shot... very sneaky
			return false, ""
		}
		// walk from one end to the other, expecting one piece exactly
		hopped := 0
		if moveRank != 0 {
			for i := math.Min(float64(srcRank), float64(tgtRank)) + 1; i < math.Max(float64(srcRank), float64(tgtRank)); i++ {
				if g.gameState.KnownBoard[int(i)][srcFile] != "." {
					hopped++
				}
			} // Can SOMEONE explain to me the need for all the casting???
		} else {
			for i := math.Min(float64(srcFile), float64(tgtFile)) + 1; i < math.Max(float64(srcFile), float64(tgtFile)); i++ {
				if g.gameState.KnownBoard[srcRank][int(i)] != "." {
					hopped++
				}
			}
		}
		if hopped != 1 {
			return false, ""
		}
	}

	// at this point you should be able to make a move... I hope
	g.gameState.KnownBoard[srcRank][srcFile], g.gameState.KnownBoard[tgtRank][tgtFile] = ".", srcPiece
	// add the target piece to dead piece if it was not an empty square
	if tgtPiece != "." {
		g.gameState.DeadPieces = append(g.gameState.DeadPieces, tgtPiece)
		g.lastDead = tgtPiece
	}

	return true, tgtPiece
}

func (g *Game) checkVictory() (victor *player, won bool) {
	defer func() {
		fmt.Printf("Game over =%v because:\n", won)
		fmt.Printf("remainingPieces: %v\n", g.gameState.RemainingPieces)
	}()
	// is black out of pieces?
	blackRemains, redRemains := false, false
	for i := 0; i < 4 && !(blackRemains && redRemains); i++ {
		for j := 0; j < 8 && !(blackRemains && redRemains); j++ {
			p := g.gameState.KnownBoard[i][j]
			redRemains = redRemains || isRed(p)
			blackRemains = blackRemains || isBlack(p)
			//fmt.Printf("Examined {%v}.  Red: %v Black %v\n", p, isRed(p), isBlack(p))
		}
	}

	for _, p := range g.gameState.RemainingPieces {
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

func (g *Game) currentPlayerOwns(rank, file int) bool {
	switch g.gameState.KnownBoard[rank][file] {
	case "K", "G", "E", "C", "H", "P", "Q":
		return g.CurrentPlayer == g.black
	case "k", "g", "e", "c", "h", "p", "q":
		return g.CurrentPlayer == g.red
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

var taunts = []string{
	"I think you should resign.",
	"This just isn't your game.",
	"Do you still think you can win?",
	"Stop! He's already dead!"}
