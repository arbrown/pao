package gamestate

// Gamestate represents the current state of a game of ban qi
type Gamestate struct {
	KnownBoard      [][]string
	RemainingPieces []string
	DeadPieces      []string
}
