package game

type command struct {
	Action, Argument string
}

type playerCommand struct {
	c command
	p *player
}

type chatCommand struct {
	Action, Player, Color, Message string
	Auth                           bool
}

type boardCommand struct {
	Action   string
	Board    [][]string
	Dead     []string
	LastMove []string
	LastDead string
	YourTurn bool
}

type colorCommand struct {
	Action, Color string
}

type gameOverCommand struct {
	Action, Message string
	YouWin          bool
}
