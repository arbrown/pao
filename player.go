package main

import (
	"github.com/gorilla/websocket"
)

func newPlayer(c *websocket.Conn, g *game) *player {
	return &player{
		ws: c,
		g:  g,
	}
}

type player struct {
	name string

	g  *game
	ws *websocket.Conn
}
