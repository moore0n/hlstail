package session

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

// Session tracks the current terminal session.
type Session struct {
	PreviousState *terminal.State
	StdinFd       int
	Paused        bool
}

// NewSession creates a new session
func NewSession() *Session {
	return &Session{}
}

// MakeRaw sets the terminal in raw mode.
func (s *Session) MakeRaw() error {
	// Get the file descriptor that is pointing to stdin.
	fd := int(os.Stdin.Fd())

	// Set to raw mode
	previousState, err := terminal.MakeRaw(fd)

	if err != nil {
		return err
	}

	s.PreviousState = previousState
	s.StdinFd = fd

	return nil
}

// GetCliWidth returns the available screen space
func (s *Session) GetCliWidth() int {
	width, _, err := terminal.GetSize(s.StdinFd)

	if err != nil {
		s.End()
		os.Exit(1)
	}

	return width
}

// End returns the terminal to its state from before hlstail started.
func (s *Session) End() {
	fmt.Print("\033[1;1H\033[0J\033[?25h")
	terminal.Restore(s.StdinFd, s.PreviousState)
}

// Start will setup the stdout for printing.
func (s *Session) Start() {
	fmt.Print("\033[1;1H\033[0J\033[?25l")
}
