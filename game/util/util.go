package util

import (
	"strconv"

	"github.com/arbrown/pao/game/command"
	"github.com/arbrown/pao/game/gamestate"
)

// ParseGameState returns a game state object given a BoardCommand
func ParseGameState(bc command.BoardCommand) gamestate.Gamestate {
	gs := gamestate.Gamestate{
		KnownBoard:      bc.Board,
		DeadPieces:      bc.Dead,
		RemainingPieces: findRemaining(bc.Board, bc.Dead),
	}
	return gs
}

// ToNotation returns ban qi notation for rank/file in integer coordinates
func ToNotation(rank, file int) string {
	return string(rune(file+'A')) + strconv.Itoa(rank+1)
}

func findRemaining(board [][]string, dead []string) []string {
	remainingPieces := []string{
		"K", "k",
		"G", "G", "g", "g",
		"E", "E", "e", "e",
		"C", "C", "c", "c",
		"H", "H", "h", "h",
		"P", "P", "P", "P", "P", "Q", "Q",
		"p", "p", "p", "p", "p", "q", "q",
	}
	for _, rank := range board {
		for _, piece := range rank {
			if piece == "." || piece == "?" {
				continue
			}
			for i, v := range remainingPieces {
				if v == piece {
					remainingPieces = append(remainingPieces[:i], remainingPieces[i+1:]...)
					break
				}
			}
		}
	}

	for _, piece := range dead {
		for i, v := range remainingPieces {
			if v == piece {
				remainingPieces = append(remainingPieces[:i], remainingPieces[i+1:]...)
				break
			}
		}
	}
	return remainingPieces
}
