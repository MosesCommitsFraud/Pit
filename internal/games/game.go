// Package games defines the common contract every casino game implements and
// the messages they use to talk to the root model. It is intentionally a leaf
// package (games import it; it imports no game) to avoid an import cycle: the
// app package assembles the registry by importing each game package.
package games

import (
	tea "github.com/charmbracelet/bubbletea"
	"pit/internal/econ"
)

// SettleMsg is emitted by a game when a wager resolves. Delta is the net change
// to the bankroll (negative for a loss, positive for a net win). The root model
// applies it and persists in career mode.
type SettleMsg struct {
	Delta int64
}

// QuitToMenuMsg asks the root model to return to the main menu.
type QuitToMenuMsg struct{}

// Settle is a convenience constructor for the settle command.
func Settle(delta int64) tea.Cmd {
	return func() tea.Msg { return SettleMsg{Delta: delta} }
}

// QuitToMenu is the command returning the player to the menu.
func QuitToMenu() tea.Msg { return QuitToMenuMsg{} }

// NewFunc builds a game model bound to the shared bankroll at the given size.
type NewFunc func(b *econ.Bankroll, width, height int) tea.Model

// Entry is one selectable game in the menu/registry.
type Entry struct {
	ID    string
	Title string
	Blurb string
	New   NewFunc
}
