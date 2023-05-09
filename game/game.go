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
	"github.com/arbrown/pao/game/command"
	"github.com/arbrown/pao/game/gamestate"
	"github.com/arbrown/pao/game/player"
	"github.com/arbrown/pao/settings"
	"github.com/gorilla/websocket"
)

// Game is a struct that represents the state and connections of a multiplayer
// ban qi game that the server is hosting.
type Game struct {
	ID                        string
	gameState                 gamestate.Gamestate
	black, red                *player.Player
	lastMove                  []string
	lastDead                  string
	active                    bool
	commandChan               chan command.PlayerCommand
	db                        *sql.DB
	CurrentPlayer, NextPlayer *player.Player
	pieceToInt                map[string]int
	canAttack                 [][]bool
	gameOverChan              chan bool
	removeGameChan            chan *Game
	kibitzers                 []*player.Player
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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Err = %v\n", err.Error())
		return false
	}
	return g.JoinWs(conn, name, user, false)
}

// JoinWs joins a game with an existing websocket
func (g *Game) JoinWs(conn *websocket.Conn, name string, user *httpauth.UserData, bot bool) bool {
	if g.CurrentPlayer == nil {
		g.CurrentPlayer = player.NewPlayer(conn, name, user, false, bot)
		fmt.Printf("User [%s] Joined as #1\n", user.Username)
		go g.listenPlayer(g.CurrentPlayer)
		go g.startGame()
		return true
	} else if g.NextPlayer == nil {
		fmt.Printf("User [%s] Trying to join as #2\n", user.Username)
		g.NextPlayer = player.NewPlayer(conn, name, user, false, bot)
		go g.listenPlayer(g.NextPlayer)
		return true
	} else {
		return g.JoinKibitz(conn, name, user)
	}
}

// JoinKibitz will create a new 'player' and add to the group of kibitzers in a game
func (g *Game) JoinKibitz(conn *websocket.Conn, name string, user *httpauth.UserData) bool {
	fmt.Printf("User [%s] Trying to join as kibitzer\n", user.Username)
	kibitzer := player.NewPlayer(conn, name, user, true, false)
	g.kibitzers = append(g.kibitzers, kibitzer)
	go g.listenPlayer(kibitzer)
	return true
}

// JoinAi will cause the specified AI to join the game
func (g *Game) JoinAi(ai settings.AiConfig) bool {
	header := http.Header{"Accept-Encoding": []string{"gzip"}}

	dialer := websocket.Dialer{
		Subprotocols:    []string{},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	fmt.Printf("Dialing AI: %v\n", ai.Address)
	conn, _, err := dialer.Dial(ai.Address, header)

	if err != nil {
		fmt.Printf("Could not dial AI successfully: %v\n", err.Error())
		return false
	}
	fmt.Println("AI dialed successfully")

	g.JoinWs(conn, ai.Name, nil, true)
	return true
}

func (g *Game) removeKibitzer(p *player.Player) {
	var newKibitzers []*player.Player
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
		g.red.Ws.Close()
	}
	if g.black != nil {
		g.black.Ws.Close()
	}
	for _, k := range g.kibitzers {
		go readLoop(k.Ws)
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

func (g *Game) handleCommand(c command.PlayerCommand) {
	fmt.Printf("Got command: %+v\n", c)
	// if g.currentPlayer != c.p {
	// 	r := command{Action: "info", Argument: "Not your turn!"}
	// 	c.p.ws.WriteJSON(r)
	// 	return
	// }
	switch c.C.Action {
	case "chat":
		color := "black"
		if c.P == g.red {
			color = "red"
		}
		if c.P.Kibitzer == true {
			color = "teal"
		}
		g.broadcastChat(c.P, c.C.Argument, color)
	case "board?":
		g.broadcastBoard()
	case "move":
		if c.P.Kibitzer == true {
			g.suggestMove(c)
		} else if ok := g.tryMove(c); ok {
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
		g.resign(c.P)
	}
}

func (g *Game) resign(p *player.Player) {
	if p.Kibitzer {
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

func (g *Game) suggestMove(pc command.PlayerCommand) bool {
	move, err := parseMove(pc.C.Argument)
	if err != nil {
		fmt.Printf("Couldn't parse move: %v\n", err.Error())
		return false
	}
	if !(move.isFlip || g.validateMove(move)) {
		return false
	}
	fmt.Printf("Move %v\n", move)
	fmt.Printf("Valid: %v", move.isValid())

	g.broadcastChat(pc.P, pc.C.Argument, "darkcyan")
	g.broadcastSuggestion(pc.P, pc.C.Argument)
	return true
}

func (g *Game) broadcastSuggestion(from *player.Player, suggestedMove string) {
	suggestion := command.SuggestCommand{Action: "suggest", Move: suggestedMove, Suggester: from.Name, Auth: from.User != nil && from.User.Username == from.Name}
	g.broadcast(suggestion)
}

func (g *Game) suggestResign(p *player.Player) {
	g.broadcastChat(p, g.getTaunt(), "darkcyan")
}

func (g *Game) getTaunt() string {
	return taunts[rand.Intn(len(taunts))]
}

func (g *Game) broadcastBoard() {
	numPlayers := 0
	if g.CurrentPlayer != nil {
		numPlayers++
	}
	if g.NextPlayer != nil {
		numPlayers++
	}
	r := command.BoardCommand{
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
		g.CurrentPlayer.Ws.WriteJSON(r)
	}
	if g.NextPlayer != nil {
		r.YourTurn = false
		g.NextPlayer.Ws.WriteJSON(r)
	}
	for _, k := range g.kibitzers {
		k.Ws.WriteJSON(r)
	}
}

func (g *Game) broadcastChat(from *player.Player, message, color string) {
	fmt.Printf("Chat from: %+v\n", from)
	chat := command.ChatCommand{Action: "chat", Player: from.Name, Message: message, Color: color, Auth: from.User != nil && from.User.Username == from.Name}
	//b, _ := json.Marshal(chat)
	//r := command{Action: "chat", Argument: string(b)}
	g.broadcast(chat)
}

func (g *Game) broadcastVictory(victor *player.Player) {
	fmt.Printf("I think the victor is: %+v\n", victor)
	var loser *player.Player
	var winColor string
	if g.black == victor {
		loser = g.red
		winColor = "black"
	} else {
		loser = g.black
		winColor = "red"
	}
	if victor != nil && victor.Ws != nil {
		fmt.Println("I told the victor he won")
		c := command.GameOverCommand{Action: "gameover", Message: "You win!", YouWin: true}
		victor.Ws.WriteJSON(c)
	}
	lose := command.GameOverCommand{Action: "gameover", Message: "You lose!", YouWin: false}
	fmt.Printf("Victor: %+v, red: %+v, black: %+v\n", victor, g.red, g.black)
	if g.red == victor && g.black != nil {
		loser = g.black
		fmt.Println("I told black he lost")
		g.black.Ws.WriteJSON(lose)
	}
	if g.black == victor && g.red != nil {
		fmt.Println("I told red he lost")
		g.red.Ws.WriteJSON(lose)
	}
	c := command.GameOverCommand{Action: "gameover", Message: "Game Over!", YouWin: false}
	for _, k := range g.kibitzers {
		k.Ws.WriteJSON(c)
	}
	g.reportVictory(victor, loser, winColor)
}

func (g *Game) reportVictory(victor, loser *player.Player, winColor string) {
	var (
		victorName, loserName string
	)
	if victor.User != nil {
		victorName = victor.User.Username
	}
	if loser.User != nil {
		loserName = loser.User.Username
	}

	if victor.Bot || loser.Bot {
		return
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
		g.CurrentPlayer.Ws.WriteJSON(v)
	}
	if g.NextPlayer != nil {
		g.NextPlayer.Ws.WriteJSON(v)
	}
	for _, k := range g.kibitzers {
		k.Ws.WriteJSON(v)
	}
}

func (g *Game) broadcastColors() {
	color := command.ColorCommand{Action: "color", Color: "red"}
	fmt.Printf("Want to broadcast colors to:\n%+v\n%+v\n", g.red, g.black)
	if g.red != nil && g.red.Ws != nil {
		g.red.Ws.WriteJSON(color)
	}
	color.Color = "black"
	if g.black != nil && g.black.Ws != nil {
		g.black.Ws.WriteJSON(color)
	}
}

func (g *Game) listenPlayer(p *player.Player) {
	fmt.Printf("Listening to new player {%+v} in game %s\n", p, g.ID)
	for {
		var com command.Command
		err := p.Ws.ReadJSON(&com)
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
			pc := command.PlayerCommand{C: com, P: p}
			g.commandChan <- pc
			// if wserr := p.ws.WriteJSON(com); wserr != nil {
			// 	fmt.Printf("Error sending JSON: %v\n", wserr.Error())
			// }
		}
	}
	fmt.Println("Stopping listen loop")
	go readLoop(p.Ws)
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
		commandChan:  make(chan command.PlayerCommand),
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

func (g *Game) tryMove(pc command.PlayerCommand) bool {
	if pc.P != g.CurrentPlayer {
		return false
	}
	move, err := parseMove(pc.C.Argument)
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

func (g *Game) validateMove(m *move) bool {
	srcFile, srcRank, tgtFile, tgtRank := m.getCoords()
	if !g.currentPlayerOwns(srcRank, srcFile) || g.currentPlayerOwns(tgtRank, tgtFile) {
		return false
	}

	srcPiece, tgtPiece := g.gameState.KnownBoard[srcRank][srcFile], g.gameState.KnownBoard[tgtRank][tgtFile]
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
			return false
		}
	}
	return true
}
func (g *Game) performMove(m *move) (bool, string) {

	if valid := g.validateMove(m); !valid {
		return valid, ""
	}
	srcFile, srcRank, tgtFile, tgtRank := m.getCoords()
	srcPiece, tgtPiece := g.gameState.KnownBoard[srcRank][srcFile], g.gameState.KnownBoard[tgtRank][tgtFile]

	// at this point you should be able to make a move... I hope
	g.gameState.KnownBoard[srcRank][srcFile], g.gameState.KnownBoard[tgtRank][tgtFile] = ".", srcPiece
	// add the target piece to dead piece if it was not an empty square
	if tgtPiece != "." {
		g.gameState.DeadPieces = append(g.gameState.DeadPieces, tgtPiece)
		g.lastDead = tgtPiece
	}

	return true, tgtPiece
}

func (g *Game) checkVictory() (victor *player.Player, won bool) {
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
	"Good heavens, are you still trying to win?",
	"Stop! He's already dead!",
	"Instead of watching this, I think I'll go find a lake of piranhas to jump in.",
	"Ban Chi is _supposed_ to be fun.",
	"Please, for the love of all that's good and holy, resign.",
	"Quitting is a temporary setback. Pain like this lasts forever.",
	"Quitting is not giving up, it's choosing to focus on something more interesting.",
	"Success in life is a slow process. Sometimes quitting what you're doing now speeds it up.",
	"Quitting is not a problem. Not knowing when to quit, is.",
	"Quitting is better than deluding yourself that you can win.",
	"Quitting is choosing happiness over pain.",
	"You'll never find a good game if you can't let go of the bad ones.",
	"You are the only one that can decide how long you will walk in hell.",
}
