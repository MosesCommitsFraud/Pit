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

// Stage centers a content block inside a fixed-size area (contentW x contentH),
// wraps it in the panel, and centers that panel in the space below the header.
// Because the inner area is a constant size, changes to the content (longer bet
// text, a status line appearing, more cards) re-center within fixed bounds
// instead of resizing the panel — which is what eliminates layout jitter.
func Stage(termW, termH, contentW, contentH int, body string) string {
	block := lipgloss.Place(contentW, contentH, lipgloss.Center, lipgloss.Center, body)
	panel := Panel.Render(block)
	return lipgloss.Place(maxi(termW, 1), maxi(termH-4, 1),
		lipgloss.Center, lipgloss.Center, panel)
}

// Reserve pads s to exactly h lines so optional content never shifts the layout.
func Reserve(h int, s string) string {
	return lipgloss.NewStyle().Height(h).Render(s)
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
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
