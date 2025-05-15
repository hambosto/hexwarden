package ui

import (
	"github.com/inancgumus/screen"
)

type Terminal struct{}

func NewTerminal() *Terminal {
	return &Terminal{}
}

func (t *Terminal) Clear() {
	screen.Clear()
}

func (t *Terminal) MoveTopLeft() {
	screen.MoveTopLeft()
}
