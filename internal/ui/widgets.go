package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"pit/internal/econ"
)

// Header renders the top bar: an inverted title chip flanked by crimson blocks
// on the left, mode and balance on the right.
func Header(game string, b *econ.Bankroll, width int) string {
	edge := AccentText
	chip := edge.Render("▌") + Chip.Render("PIT · "+Caps(game)) + edge.Render("▐")
	right := Subtle.Render(Caps(b.Mode().String())+"   ") + Heading.Render(Money(b.Balance()))
	gap := width - lipgloss.Width(chip) - lipgloss.Width(right) - 1
	if gap < 1 {
		gap = 1
	}
	return " " + chip + strings.Repeat(" ", gap) + right
}

// Rule draws a full-width double-line divider.
func Rule(width int) string {
	if width < 1 {
		width = 1
	}
	return RuleStyle.Render(strings.Repeat("═", width))
}

// Screen composes the standard layout: title bar, top rule, a fixed-height
// content area (top-left aligned so content never re-centers or jumps), bottom
// rule, and a help line. Top-alignment plus reserved-height regions inside the
// body are what keep the layout stable.
func Screen(title string, b *econ.Bankroll, width, height int, body, hints string) string {
	header := Header(title, b, width)
	contentH := height - 4
	if contentH < 1 {
		contentH = 1
	}
	content := lipgloss.NewStyle().Padding(1, 2).Height(contentH).Render(body)
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		Rule(width),
		content,
		Rule(width),
		" "+HelpBar.Render(hints),
	)
}

// Reserve pads s to exactly h lines so optional content never shifts layout.
func Reserve(h int, s string) string {
	return lipgloss.NewStyle().Height(h).Render(s)
}

// SectionLabel renders an uppercase section heading, optionally with a value.
func SectionLabel(name, value string) string {
	if value == "" {
		return Label.Render(Caps(name))
	}
	return Label.Render(Caps(name)) + "  " + Heading.Render(value)
}

// Result is the outcome of a settled wager, for the banner color.
type Result int

const (
	ResultNeutral Result = iota
	ResultWin
	ResultLose
	ResultPush
)

// Banner renders a colored outcome line, e.g. "WIN  +$200".
func Banner(r Result, label string, delta int64) string {
	switch r {
	case ResultWin:
		return WinText.Render(fmt.Sprintf("%s  +%s", Caps(label), Money(delta)))
	case ResultLose:
		return LoseText.Render(fmt.Sprintf("%s  %s", Caps(label), Money(delta)))
	case ResultPush:
		return Subtle.Bold(true).Render(Caps(label) + "  PUSH")
	default:
		return Heading.Render(label)
	}
}

// BetSelector renders the current bet with adjustment hints.
func BetSelector(bet int64) string {
	return Label.Render("BET") + "  " +
		AccentText.Render(Money(bet)) +
		Subtle.Render("   ‹ ›")
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ = maxi
