package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"pit/internal/deck"
)

var cardFace = lipgloss.NewStyle().
	Background(Cream).
	Foreground(Ink).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#cbbf9a"))

var cardBack = lipgloss.NewStyle().
	Background(lipgloss.Color("#7a1f1f")).
	Foreground(Gold).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(Gold)

// RenderCard draws a single face-up card as a small box.
func RenderCard(c deck.Card) string {
	suit := c.Suit.Symbol()
	rank := c.Rank.Label()
	fg := Ink
	if c.Suit.Red() {
		fg = Red
	}
	style := cardFace.Foreground(fg)
	// Pad the rank so the 5-wide interior aligns for "10" and single chars.
	top := pad(rank, 5, true)
	mid := center(suit, 5)
	bot := pad(rank, 5, false)
	return style.Render(top + "\n" + mid + "\n" + bot)
}

// RenderHidden draws a face-down card.
func RenderHidden() string {
	return cardBack.Render("░░░░░\n░ ? ░\n░░░░░")
}

// RenderHand lays cards out left to right; hideFirst hides the hole card.
func RenderHand(cards []deck.Card, hidden int) string {
	boxes := make([]string, 0, len(cards)+hidden)
	for _, c := range cards {
		boxes = append(boxes, RenderCard(c))
	}
	for i := 0; i < hidden; i++ {
		boxes = append(boxes, RenderHidden())
	}
	if len(boxes) == 0 {
		return ""
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, boxes...)
}

func pad(s string, w int, left bool) string {
	for len(s) < w {
		if left {
			s = s + " "
		} else {
			s = " " + s
		}
	}
	return s
}

func center(s string, w int) string {
	gap := w - len(s)
	if gap <= 0 {
		return s
	}
	l := gap / 2
	return strings.Repeat(" ", l) + s + strings.Repeat(" ", gap-l)
}
