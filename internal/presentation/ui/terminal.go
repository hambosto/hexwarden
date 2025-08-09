package ui

import (
	"fmt"

	"github.com/inancgumus/screen"
)

// Terminal provides terminal control functionality
type Terminal struct{}

// NewTerminal creates a new Terminal instance
func NewTerminal() *Terminal {
	return &Terminal{}
}

// Clear clears the terminal screen
func (t *Terminal) Clear() {
	screen.Clear()
}

// MoveTopLeft moves the cursor to the top-left corner
func (t *Terminal) MoveTopLeft() {
	screen.MoveTopLeft()
}

// ShowCursor shows the terminal cursor
func (t *Terminal) ShowCursor() {
	fmt.Print("\033[?25h")
}

// SetTitle sets the terminal window title
func (t *Terminal) SetTitle(title string) {
	fmt.Printf("\033]0;%s\007", title)
}

// PrintBanner prints the application banner
func (t *Terminal) PrintBanner() {
	banner := `
██╗  ██╗███████╗██╗  ██╗██╗    ██╗ █████╗ ██████╗ ██████╗ ███████╗███╗   ██╗
██║  ██║██╔════╝╚██╗██╔╝██║    ██║██╔══██╗██╔══██╗██╔══██╗██╔════╝████╗  ██║
███████║█████╗   ╚███╔╝ ██║ █╗ ██║███████║██████╔╝██║  ██║█████╗  ██╔██╗ ██║
██╔══██║██╔══╝   ██╔██╗ ██║███╗██║██╔══██║██╔══██╗██║  ██║██╔══╝  ██║╚██╗██║
██║  ██║███████╗██╔╝ ██╗╚███╔███╔╝██║  ██║██║  ██║██████╔╝███████╗██║ ╚████║
╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚═╝  ╚═══╝
                                                                              
                    Secure File Encryption Tool v1.1.0
`
	fmt.Println(banner)
}

// PrintSeparator prints a visual separator
func (t *Terminal) PrintSeparator() {
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
}

// PrintSuccess prints a success message with formatting
func (t *Terminal) PrintSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// PrintError prints an error message with formatting
func (t *Terminal) PrintError(message string) {
	fmt.Printf("❌ %s\n", message)
}

// PrintWarning prints a warning message with formatting
func (t *Terminal) PrintWarning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// PrintInfo prints an info message with formatting
func (t *Terminal) PrintInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// Cleanup performs terminal cleanup operations
func (t *Terminal) Cleanup() {
	t.ShowCursor()
	fmt.Println() // Add a newline for clean exit
}
