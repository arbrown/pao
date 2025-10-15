package player

import (
	"io"

	"github.com/apexskier/httpauth"
)

// WsConn is an interface that abstracts websocket operations for testing
type WsConn interface {
	WriteJSON(v interface{}) error
	ReadJSON(v interface{}) error
	Close() error
	NextReader() (messageType int, r io.Reader, err error)
}

// NewPlayer will create a new player (or kibitzer) for a pao server
func NewPlayer(c WsConn, name string, user *httpauth.UserData, kibitzer bool, bot bool) *Player {
	return &Player{
		Ws:        c,
		Name:      name,
		User:      user,
		Kibitzer:  kibitzer,
		Bot:       bot,
		SessionID: "", // Will be set by game when needed
	}
}

// NewPlayerWithSession creates a player with a specific session ID (for reconnection)
func NewPlayerWithSession(c WsConn, name string, user *httpauth.UserData, kibitzer bool, bot bool, sessionID string) *Player {
	return &Player{
		Ws:        c,
		Name:      name,
		User:      user,
		Kibitzer:  kibitzer,
		Bot:       bot,
		SessionID: sessionID,
	}
}

// Player contains information about a player and how to communicate with them
type Player struct {
	Name      string
	Ws        WsConn
	User      *httpauth.UserData
	Kibitzer  bool
	Bot       bool
	SessionID string // Unique session ID for reconnection
}
