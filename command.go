package main

type command struct {
	Action, Argument string
}

type playerCommand struct {
	c command
	p *player
}

type chatCommand struct {
	Action, Player, Message string
}

type boardCommand struct {
	Action   string
	Board    [][]string
	YourTurn bool
}

type colorCommand struct {
	Action, Color string
}

type gameOverCommand struct {
	Action, Message string
	YouWin          bool
}
