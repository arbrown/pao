package main

type command struct {
	Action, Argument string
}

type playerCommand struct {
	c command
	p *player
}

type chatMessage struct {
	Player, Message string
}
