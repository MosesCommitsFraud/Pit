package blackjack

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"pit/internal/deck"
	"pit/internal/ui"
)

func (m *Model) View() string {
	header := ui.Header("Blackjack", m.bank, m.width)

	dealer := m.viewDealer()
	player := m.viewPlayers()

	var status string
	switch m.phase {
	case phaseBetting:
		status = ui.Subtle.Render("Place your bet · enter to deal")
		if m.outcome != "" {
			status = ui.LoseText.Render(m.outcome)
		}
	case phasePlayer:
		status = ui.Heading.Render("Your move")
	case phaseResult:
		if m.delta > 0 {
			status = ui.Banner(ui.ResultWin, m.outcome, m.delta)
		} else if m.delta < 0 {
			status = ui.Banner(ui.ResultLose, m.outcome, m.delta)
		} else {
			status = ui.Banner(ui.ResultPush, m.outcome, 0)
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		ui.Subtle.Render("DEALER"),
		dealer,
		"",
		ui.Subtle.Render("YOU"),
		player,
		"",
		ui.BetSelector(m.bet()),
		"",
		status,
	)

	center := lipgloss.Place(max(m.width, 1), max(m.height-4, 1),
		lipgloss.Center, lipgloss.Center, ui.Panel.Render(body))

	return lipgloss.JoinVertical(lipgloss.Left, header, center, m.helpLine())
}

func (m *Model) helpLine() string {
	switch m.phase {
	case phasePlayer:
		h := &m.players[m.active]
		hints := "h hit · s stand"
		if len(h.cards) == 2 {
			hints += " · d double"
			if m.canSplit(h) {
				hints += " · p split"
			}
		}
		return ui.Help(hints + " · m menu")
	case phaseResult:
		return ui.Help("enter deal again · m menu")
	default:
		return ui.Help("←/→ bet · enter deal · m menu")
	}
}

func (m *Model) viewDealer() string {
	_, dShown := m.visibleCounts()
	cards := m.dealer
	if dShown < len(cards) {
		cards = cards[:dShown]
	}
	hidden := 0
	shown := cards
	if m.holeDown && len(cards) >= 2 {
		// keep the first card up, hole card face-down
		shown = cards[:1]
		hidden = 1
	}
	row := ui.RenderHand(shown, hidden)
	label := ""
	if !m.holeDown && len(m.dealer) > 0 {
		v, _ := handValue(m.dealer)
		label = "  " + ui.Heading.Render(valueLabel(m.dealer, v))
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, label)
}

func (m *Model) viewPlayers() string {
	pShown, _ := m.visibleCounts()
	cols := make([]string, 0, len(m.players))
	for i, h := range m.players {
		cards := h.cards
		if m.phase == phaseDealing && i == 0 && pShown < len(cards) {
			cards = cards[:pShown]
		}
		row := ui.RenderHand(cards, 0)
		v, _ := handValue(cards)
		tag := valueLabel(cards, v)
		if h.bust {
			tag = ui.LoseText.Render("BUST " + tag)
		} else if h.doubled {
			tag += " ×2"
		}
		col := lipgloss.JoinVertical(lipgloss.Left, row, ui.Subtle.Render(tag))
		if m.phase == phasePlayer && i == m.active && len(m.players) > 1 {
			col = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).
				BorderForeground(ui.Gold).Padding(0, 1).Render(col)
		}
		cols = append(cols, col)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

// visibleCounts returns how many player / dealer cards to show mid-deal.
func (m *Model) visibleCounts() (player, dealer int) {
	if m.phase != phaseDealing {
		return 99, 99
	}
	// reveal order: player, dealer, player, dealer
	switch m.reveal {
	case 0:
		return 0, 0
	case 1:
		return 1, 0
	case 2:
		return 1, 1
	case 3:
		return 2, 1
	default:
		return 2, 2
	}
}

func valueLabel(cards []deck.Card, v int) string {
	_, soft := handValue(cards)
	if len(cards) == 2 && v == 21 {
		return "Blackjack!"
	}
	if soft && v <= 21 {
		return fmt.Sprintf("soft %d", v)
	}
	return fmt.Sprintf("%d", v)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
