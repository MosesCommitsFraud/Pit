package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"pit/internal/deck"
)

var (
	bracket   = lipgloss.NewStyle().Foreground(Faint)
	blackSuit = lipgloss.NewStyle().Bold(true).Foreground(Bright)
	redSuit   = lipgloss.NewStyle().Bold(true).Foreground(Accent)
	rankStyle = lipgloss.NewStyle().Bold(true).Foreground(Bright)
	backStyle = lipgloss.NewStyle().Foreground(Dim)
)

// RenderCard draws one face-up card as a fixed-width bracket token, e.g.
// "[ 10♠ ]". Width is constant (7 cols) so rows of cards always align.
func RenderCard(c deck.Card) string {
	rank := rankStyle.Render(fmt.Sprintf("%2s", c.Rank.Label()))
	ss := blackSuit
	if c.Suit.Red() {
		ss = redSuit
	}
	return bracket.Render("[ ") + rank + ss.Render(c.Suit.Symbol()) + bracket.Render(" ]")
}

// RenderHidden draws a face-down card, same width as a face card.
func RenderHidden() string {
	return bracket.Render("[ ") + backStyle.Render("░░░") + bracket.Render(" ]")
}

// RenderEmpty draws an empty card slot (community placeholder).
func RenderEmpty() string {
	return bracket.Render("[ ") + RuleStyle.Render("   ") + bracket.Render(" ]")
}

// RenderHand lays cards out left to right; `hidden` appends face-down cards.
func RenderHand(cards []deck.Card, hidden int) string {
	tokens := make([]string, 0, len(cards)+hidden)
	for _, c := range cards {
		tokens = append(tokens, RenderCard(c))
	}
	for i := 0; i < hidden; i++ {
		tokens = append(tokens, RenderHidden())
	}
	if len(tokens) == 0 {
		return ""
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, joinSpaced(tokens)...)
}

// joinSpaced inserts a single space between tokens.
func joinSpaced(tokens []string) []string {
	out := make([]string, 0, len(tokens)*2)
	for i, t := range tokens {
		if i > 0 {
			out = append(out, " ")
		}
		out = append(out, t)
	}
	return out
}
