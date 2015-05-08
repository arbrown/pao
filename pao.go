package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func main() {
	fmt.Printf("Hello, Pao\n")
	removeGameChan := make(chan *game)
	games := make(map[string]*game)
	http.Handle("/game", gameHandler{games: games, removeGameChan: removeGameChan})
	http.Handle("/listGames", listGamesHandler{games: games})
	http.Handle("/", http.FileServer(http.Dir(".")))

	go func() {
		for {
			select {
			case g := <-removeGameChan:
				delete(games, g.id)
				break
			}
		}
	}()

	err := http.ListenAndServe(":2015", nil)
	if err != nil {
		panic("ListenAndServe:" + err.Error())
	}
}

type gameHandler struct {
	games          map[string]*game
	removeGameChan chan *game
}

type listGamesHandler struct {
	games map[string]*game
}

type gameResponse struct {
	ID      string
	Players []string
}

func (lgh listGamesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var resp []gameResponse
	for id, game := range lgh.games {
		var players []string
		if game.currentPlayer != nil {
			players = append(players, game.currentPlayer.name)
		}
		if game.nextPlayer != nil {
			players = append(players, game.nextPlayer.name)
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
		if game, ok := gh.games[id]; ok {
			fmt.Println("Trying to join existing game")
			game.join(w, r, name)
		} else {
			// make the id requested
			g := newGame(id, gh.removeGameChan)
			fmt.Printf("Made new game %s\n", id)
			gh.games[g.id] = g
			g.join(w, r, name)
		}
	} else {
		// no id specified, make the game
		newID := 0
		for _, exists := gh.games[strconv.Itoa(newID)]; exists; _, exists = gh.games[strconv.Itoa(newID)] {
			newID++
		}
		g := newGame(strconv.Itoa(newID), gh.removeGameChan)
		fmt.Printf("Made new game %d\n", newID)
		gh.games[g.id] = g
		g.join(w, r, name)
	}
}
