package game

import (
	"github.com/gorilla/websocket"
)

func newPlayer(c *websocket.Conn, g *Game, name string) *player {
	return &player{
		ws:   c,
		g:    g,
		Name: name,
	}
}

type player struct {
	Name string

	g  *Game
	ws *websocket.Conn
}
