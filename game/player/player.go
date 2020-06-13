package player

import (
	"github.com/apexskier/httpauth"
	"github.com/gorilla/websocket"
)

// NewPlayer will create a new player (or kibitzer) for a pao server
func NewPlayer(c *websocket.Conn, name string, user *httpauth.UserData, kibitzer bool, bot bool) *Player {
	return &Player{
		Ws:       c,
		Name:     name,
		User:     user,
		Kibitzer: kibitzer,
		Bot:      bot,
	}
}

// Player contains information about a player and how to communicate with them
type Player struct {
	Name     string
	Ws       *websocket.Conn
	User     *httpauth.UserData
	Kibitzer bool
	Bot      bool
}
