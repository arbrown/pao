package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	paoDb "github.com/arbrown/pao/db"
	"github.com/arbrown/pao/game"
	"github.com/arbrown/pao/settings"
	_ "github.com/mattn/go-sqlite3"
)

var (
	a *paoDb.Auth
)

func main() {
	fmt.Printf("Hello, Pao\n")
	removeGameChan := make(chan *game.Game)
	games := make(map[string]*game.Game)
	s, err := settings.GetSettings()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("Settings: %+v\n", s)
	}
	// create db connection
	db, err := sql.Open(s.DbConfig.Driver, s.DbConfig.ConnectionString)
	if err != nil {
		db = nil
		fmt.Printf("Could not open database: %s\n", err)
	} else {
		paoDb.Init(db)
	}
	a, err = paoDb.NewAuth(s)
	if err != nil {
		fmt.Println("Could not create auth")
		fmt.Println(err.Error())
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/listGames", listGamesHandler{games: games})
	httpMux.Handle("/game", gameHandler{games: games, removeGameChan: removeGameChan, db: db})
	httpMux.Handle("/playAi", playAiHandler{games: games, removeGameChan: removeGameChan, db: db, ais: s.Ais})
	httpMux.HandleFunc("/login", a.PostLogin)
	httpMux.HandleFunc("/register", a.PostRegister)
	httpMux.HandleFunc("/logout", a.HandleLogout)
	httpMux.HandleFunc("/cu", a.Cu)
	httpMux.HandleFunc("/leaderBoard", paoDb.GetLeaderBoard)

	// Serve the UI here
	httpMux.Handle("/", http.FileServer(http.Dir("./react-app/build/")))

	go func() {
		for {
			select {
			case g := <-removeGameChan:
				fmt.Printf("Removing game: %v\n", g.ID)
				delete(games, g.ID)
				break
			}
		}
	}()

	host, port := os.Getenv("HOST"), os.Getenv("PORT")
	if port == "" {
		port = "2015"
	}

	bind := fmt.Sprintf("%s:%s", host, port)

	fmt.Printf("Listening on %s\n", bind)
	err = http.ListenAndServe(bind, httpMux)
	if err != nil {
		panic("ListenAndServe:" + err.Error())
	}
}

type gameHandler struct {
	games          map[string]*game.Game
	removeGameChan chan *game.Game
	db             *sql.DB
}

type playAiHandler struct {
	games          map[string]*game.Game
	removeGameChan chan *game.Game
	db             *sql.DB
	ais            []settings.AiConfig
}

type listGamesHandler struct {
	games map[string]*game.Game
}

type gameResponse struct {
	ID      string
	Players []string
}

func (lgh listGamesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var resp []gameResponse
	for id, game := range lgh.games {
		var players []string
		if game.CurrentPlayer() != nil {
			players = append(players, game.CurrentPlayer().Name)
		}
		if game.NextPlayer() != nil {
			players = append(players, game.NextPlayer().Name)
		}
		if len(players) != 0 {
			resp = append(resp, gameResponse{ID: id, Players: players})
		}
	}
	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (gh gameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var user = a.GetUser(w, r)
	fmt.Printf("105: user: %+v\n", user)
	id := r.FormValue("id")
	name := r.FormValue("name")
	fmt.Println(id)
	if name == "" {
		name = "Anonymous Coward"
	}
	if id != "" {
		// Does the game exist?
		fmt.Printf("All Games: %v\n", gh.games)
		if existingGame, ok := gh.games[id]; ok {
			fmt.Println("Trying to join existing game")
			existingGame.Join(w, r, name, user)
		} else {
			// make the id requested
			g := game.NewGame(id, gh.removeGameChan, gh.db)
			fmt.Printf("Made new game %s\n", id)
			gh.games[g.ID] = g
			g.Join(w, r, name, user)
		}
	} else {
		// no id specified, make the game
		newID := 0
		for _, exists := gh.games[strconv.Itoa(newID)]; exists; _, exists = gh.games[strconv.Itoa(newID)] {
			newID++
		}
		g := game.NewGame(strconv.Itoa(newID), gh.removeGameChan, gh.db)
		fmt.Printf("Made new game %d\n", newID)
		gh.games[g.ID] = g
		fmt.Printf("Trying to join new game as: %+v\n", user)
		if ok := g.Join(w, r, name, user); !ok {
			delete(gh.games, g.ID)
		}
	}
}

// Create a new game against an AI and return the game ID to the client
func (pah playAiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("ai")
	var ai settings.AiConfig
	for _, aic := range pah.ais {
		if aic.Name == name {
			ai = aic
		}
	}
	if ai.Name == "" {
		fmt.Printf("Couldn't find AI: '%v'\n", name)
	}

	newID := 0
	for _, exists := pah.games[strconv.Itoa(newID)]; exists; _, exists = pah.games[strconv.Itoa(newID)] {
		newID++
	}

	g := game.NewGame(strconv.Itoa(newID), pah.removeGameChan, pah.db)
	fmt.Printf("Made new game %d\n", newID)
	pah.games[g.ID] = g
	// if ok := g.Join(w, r, name, user); !ok {
	// 	delete(pah.games, g.ID)
	// }

	fmt.Printf("AI tries to join game: %v\n", g)
	if ok := g.JoinAi(ai); !ok {
		// if game join fails, remove game from list
		delete(pah.games, g.ID)
	}

	type aiGameResponse struct {
		ID string
	}

	var resp = aiGameResponse{
		ID: g.ID,
	}
	js, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("Couldn't write json response: %v\n", resp)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
