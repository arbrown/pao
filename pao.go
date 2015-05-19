package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/arbrown/pao/game"
)

func main() {
	fmt.Printf("Hello, Pao\n")
	removeGameChan := make(chan *game.Game)
	games := make(map[string]*game.Game)
	httpMux, wsMux := http.NewServeMux(), http.NewServeMux()
	httpMux.Handle("/listGames", listGamesHandler{games: games})
	httpMux.Handle("/", http.FileServer(http.Dir("./client/")))

	wsMux.Handle("/game", gameHandler{games: games, removeGameChan: removeGameChan})

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

	// open shift requires web sockets to be on this port
	wsPort := "8000"
	bind := fmt.Sprintf("%s:%s", host, port)
	wsBind := fmt.Sprintf("%s:%s", host, wsPort)

	go func() {
		fmt.Printf("Listening on %s\n", wsBind)
		http.ListenAndServe(wsBind, wsMux)
	}()

	err := http.ListenAndServe(bind, httpMux)
	if err != nil {
		panic("ListenAndServe:" + err.Error())
	}
}

type gameHandler struct {
	games          map[string]*game.Game
	removeGameChan chan *game.Game
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
		if game.CurrentPlayer != nil {
			players = append(players, game.CurrentPlayer.Name)
		}
		if game.NextPlayer != nil {
			players = append(players, game.NextPlayer.Name)
		}
		resp = append(resp, gameResponse{ID: id, Players: players})
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
			existingGame.Join(w, r, name)
		} else {
			// make the id requested
			g := game.NewGame(id, gh.removeGameChan)
			fmt.Printf("Made new game %s\n", id)
			gh.games[g.ID] = g
			g.Join(w, r, name)
		}
	} else {
		// no id specified, make the game
		newID := 0
		for _, exists := gh.games[strconv.Itoa(newID)]; exists; _, exists = gh.games[strconv.Itoa(newID)] {
			newID++
		}
		g := game.NewGame(strconv.Itoa(newID), gh.removeGameChan)
		fmt.Printf("Made new game %d\n", newID)
		gh.games[g.ID] = g
		g.Join(w, r, name)
	}
}
