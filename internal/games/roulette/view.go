package roulette

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"pit/internal/ui"
)

// pocketText renders a number colored by its pocket: crimson for red, bright
// for black, inverted for zero.
func pocketText(n int) string {
	st := lipgloss.NewStyle().Bold(true)
	switch {
	case n == 0:
		st = st.Foreground(ui.Black).Background(ui.Bright)
	case isRed(n):
		st = st.Foreground(ui.Accent)
	default:
		st = st.Foreground(ui.Bright)
	}
	return st.Render(fmt.Sprintf("%2d", n))
}

func bigNumber(n int) string {
	st := lipgloss.NewStyle().Bold(true)
	switch {
	case n == 0:
		st = st.Foreground(ui.Black).Background(ui.Bright)
	case isRed(n):
		st = st.Foreground(ui.Accent)
	default:
		st = st.Foreground(ui.Bright)
	}
	inner := st.Padding(0, 2).Render(fmt.Sprintf("%2d", n))
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).BorderForeground(ui.Faint).
		Padding(0, 1).Render(inner)
}

func (m *Model) View() string {
	board := m.viewBoard()

	chipLine := ui.Label.Render("CHIP") + "  " + ui.AccentText.Render(ui.Money(m.chip())) +
		ui.Subtle.Render(" ‹ ›    ") + ui.Label.Render("WAGERED") + "  " + ui.Heading.Render(ui.Money(m.totalWagered()))

	var status string
	switch m.phase {
	case phaseSpin:
		status = ui.Subtle.Render("spinning…")
	case phaseResult:
		if m.delta > 0 {
			status = ui.Banner(ui.ResultWin, m.outcome, m.delta)
		} else if m.delta < 0 {
			status = ui.Banner(ui.ResultLose, m.outcome, m.delta)
		} else {
			status = ui.Banner(ui.ResultPush, m.outcome, 0)
		}
	default:
		if m.outcome != "" {
			status = ui.LoseText.Render(m.outcome)
		} else {
			status = ui.Subtle.Render("place chips, then SPACE to spin")
		}
	}

	right := lipgloss.NewStyle().Width(40).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			chipLine,
			"",
			bigNumber(wheel[m.wheelPos]),
			"",
			ui.SectionLabel("recent", ""),
			ui.Reserve(1, m.viewHistory()),
			"",
			ui.Reserve(1, status),
		),
	)

	body := lipgloss.JoinHorizontal(lipgloss.Top, board, "     ", right)
	return ui.Screen("Roulette", m.bank, m.width, m.height, body, m.hints())
}

func (m *Model) viewBoard() string {
	rows := []string{ui.SectionLabel("table", "")}
	for i, t := range targets {
		name := t.label
		if t.kind == betStraight {
			name = fmt.Sprintf("Straight %d", m.number)
		}
		chip := ""
		if m.wagers[i] > 0 {
			chip = "  " + ui.AccentText.Render(ui.Money(m.wagers[i]))
		}
		line := ui.Caps(fmt.Sprintf("%-12s %2d:1", name, t.payout)) + chip
		if i == m.cursor {
			rows = append(rows, ui.Selected.Render("▸ "+line))
		} else {
			rows = append(rows, ui.Unselected.Render("  "+line))
		}
	}
	inner := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.NewStyle().Width(28).Render(inner)
}

func (m *Model) viewHistory() string {
	if len(m.history) == 0 {
		return ui.Subtle.Render("—")
	}
	shown := m.history
	if len(shown) > 8 {
		shown = shown[:8]
	}
	cells := make([]string, 0, len(shown)*2)
	for i, n := range shown {
		if i > 0 {
			cells = append(cells, " ")
		}
		cells = append(cells, pocketText(n))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, cells...)
}

func (m *Model) hints() string {
	if m.phase == phaseSpin {
		return "M menu"
	}
	if targets[m.cursor].kind == betStraight {
		return "↑↓ bet · −+ number · ‹ › chip · ENTER place · SPACE spin · M menu"
	}
	return "↑↓ bet · ‹ › chip · ENTER place · BACKSPACE clear · SPACE spin · M menu"
}
