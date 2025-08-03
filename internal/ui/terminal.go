package ui

import (
	"github.com/inancgumus/screen"
)

// Terminal provides methods for terminal screen manipulation.
type Terminal struct{}

// NewTerminal creates a new Terminal instance.
func NewTerminal() *Terminal {
	return &Terminal{}
}

// Clear clears the terminal screen.
func (t *Terminal) Clear() {
	screen.Clear()
}

// MoveTopLeft moves the cursor to the top-left corner of the terminal.
func (t *Terminal) MoveTopLeft() {
	screen.MoveTopLeft()
}
