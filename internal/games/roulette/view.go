package roulette

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"pit/internal/ui"
)

var (
	redChip   = lipgloss.NewStyle().Bold(true).Foreground(ui.Cream).Background(ui.Red)
	blackChip = lipgloss.NewStyle().Bold(true).Foreground(ui.Cream).Background(lipgloss.Color("#222"))
	greenChip = lipgloss.NewStyle().Bold(true).Foreground(ui.Ink).Background(ui.Green)
)

func pocketStyle(n int) lipgloss.Style {
	switch {
	case n == 0:
		return greenChip
	case isRed(n):
		return redChip
	default:
		return blackChip
	}
}

func renderPocket(n int) string {
	return pocketStyle(n).Padding(0, 1).Render(fmt.Sprintf("%2d", n))
}

func (m *Model) View() string {
	header := ui.Header("Roulette", m.bank, m.width)

	// big landing/spinning pocket
	current := wheel[m.wheelPos]
	bigStyle := pocketStyle(current).Bold(true).Padding(1, 3)
	big := bigStyle.Render(fmt.Sprintf("%d", current))
	wheelBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).BorderForeground(ui.Gold).Padding(0, 1).
		Render(big)

	board := m.viewBoard()

	hist := m.viewHistory()

	var status string
	switch m.phase {
	case phaseSpin:
		status = ui.Subtle.Render("Spinning…")
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
			status = ui.Subtle.Render("Place chips, then SPACE to spin")
		}
	}

	chipLine := ui.Heading.Render("Chip: ") +
		lipgloss.NewStyle().Bold(true).Foreground(ui.Gold).Render(ui.Money(m.chip())) +
		ui.Subtle.Render("  (←/→)   ") +
		ui.Subtle.Render("Total wagered: ") + ui.Heading.Render(ui.Money(m.totalWagered()))

	right := lipgloss.JoinVertical(lipgloss.Center, wheelBox, "", hist, "", status)

	body := lipgloss.JoinHorizontal(lipgloss.Top, board, "    ",
		lipgloss.JoinVertical(lipgloss.Left, chipLine, "", right))

	center := lipgloss.Place(max(m.width, 1), max(m.height-4, 1),
		lipgloss.Center, lipgloss.Center, ui.Panel.Render(body))

	return lipgloss.JoinVertical(lipgloss.Left, header, center, m.helpLine())
}

func (m *Model) viewBoard() string {
	rows := []string{ui.Heading.Render("Bets")}
	for i, t := range targets {
		label := t.label
		if t.kind == betStraight {
			label = fmt.Sprintf("Straight  %s", renderPocket(m.number))
		}
		payout := ui.Subtle.Render(fmt.Sprintf(" %d:1", t.payout))
		chip := ""
		if m.wagers[i] > 0 {
			chip = "  " + lipgloss.NewStyle().Bold(true).Foreground(ui.Gold).Render(ui.Money(m.wagers[i]))
		}
		line := label + payout + chip
		if i == m.cursor {
			rows = append(rows, ui.Selected.Render("› "+line))
		} else {
			rows = append(rows, ui.Unselected.Render("  "+line))
		}
	}
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).
		BorderForeground(ui.Border).Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (m *Model) viewHistory() string {
	if len(m.history) == 0 {
		return ui.Subtle.Render("no spins yet")
	}
	cells := make([]string, 0, len(m.history))
	for _, n := range m.history {
		cells = append(cells, renderPocket(n))
	}
	return ui.Subtle.Render("Recent: ") + lipgloss.JoinHorizontal(lipgloss.Left, cells...)
}

func (m *Model) helpLine() string {
	switch m.phase {
	case phaseResult, phaseBetting:
		hints := "↑/↓ bet · ←/→ chip · enter place · backspace clear · space spin"
		if targets[m.cursor].kind == betStraight {
			hints = "↑/↓ bet · -/+ number · ←/→ chip · enter place · space spin"
		}
		return ui.Help(hints + " · m menu")
	default:
		return ui.Help("m menu")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
