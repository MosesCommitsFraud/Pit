package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"pit/internal/econ"
)

// Header renders the persistent top bar: game title on the left, mode and
// balance on the right.
func Header(game string, b *econ.Bankroll, width int) string {
	left := Title.Render("♠ " + game)
	right := Subtle.Render(b.Mode().String()+"  ") +
		lipgloss.NewStyle().Bold(true).Foreground(Gold).Render(Money(b.Balance()))
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	bar := left + lipgloss.NewStyle().Width(gap).Render("") + right
	return lipgloss.NewStyle().
		Width(width).
		BorderBottom(true).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(Border).
		Render(bar)
}

// Help renders a dimmed key-hint line.
func Help(hints string) string { return HelpBar.Render(hints) }

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
		return WinText.Render(fmt.Sprintf("%s  +%s", label, Money(delta)))
	case ResultLose:
		return LoseText.Render(fmt.Sprintf("%s  %s", label, Money(delta)))
	case ResultPush:
		return Subtle.Bold(true).Render(label + "  push")
	default:
		return Heading.Render(label)
	}
}

// BetSelector renders the current bet with adjustment hints.
func BetSelector(bet int64) string {
	return Heading.Render("Bet: ") +
		lipgloss.NewStyle().Bold(true).Foreground(Gold).Render(Money(bet)) +
		Subtle.Render("   (←/→ adjust)")
}
