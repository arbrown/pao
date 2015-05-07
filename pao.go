package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func main() {
	fmt.Printf("Hello, Pao\n")
	games := make(map[string]*game)
	http.Handle("/game", gameHandler{games: games})
	http.Handle("/", http.FileServer(http.Dir(".")))

	err := http.ListenAndServe(":2015", nil)
	if err != nil {
		panic("ListenAndServe:" + err.Error())
	}
}

type gameHandler struct {
	games map[string]*game
}

func (gh gameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	fmt.Println(id)
	if id != "" {
		// Does the game exist?
		if game, ok := gh.games[id]; ok {
			fmt.Println("Trying to join existing game")
			game.join(w, r)
		}
	} else {
		// no id specified, make the game
		newID := 0
		for _, exists := gh.games[strconv.Itoa(newID)]; exists; _, exists = gh.games[strconv.Itoa(newID)] {
			newID++
		}
		g := newGame(strconv.Itoa(newID))
		fmt.Printf("Made new game %d\n", newID)
		gh.games[g.id] = g
		g.join(w, r)
	}
}
