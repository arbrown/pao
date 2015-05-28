package game

import (
	"github.com/apexskier/httpauth"
	"github.com/gorilla/websocket"
)

func newPlayer(c *websocket.Conn, g *Game, name string, user *httpauth.UserData) *player {
	return &player{
		ws:   c,
		g:    g,
		Name: name,
		user: user,
	}
}

type player struct {
	Name string

	g    *Game
	ws   *websocket.Conn
	user *httpauth.UserData
}
