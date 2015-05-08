package main

import (
	"github.com/gorilla/websocket"
)

func newPlayer(c *websocket.Conn, g *game, name string) *player {
	return &player{
		ws:   c,
		g:    g,
		name: name,
	}
}

type player struct {
	name string

	g  *game
	ws *websocket.Conn
}
