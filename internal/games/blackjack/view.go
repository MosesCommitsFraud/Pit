package blackjack

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"pit/internal/deck"
	"pit/internal/ui"
)

func (m *Model) View() string {
	dealerVal := ""
	if !m.holeDown && len(m.dealer) > 0 {
		v, _ := handValue(m.dealer)
		dealerVal = valueLabel(m.dealer, v)
	}

	var status string
	switch m.phase {
	case phaseBetting:
		status = ui.Subtle.Render("place your bet · ENTER to deal")
		if m.outcome != "" {
			status = ui.LoseText.Render(m.outcome)
		}
	case phasePlayer:
		status = ui.AccentText.Render("YOUR MOVE")
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
		ui.SectionLabel("dealer", dealerVal),
		ui.Reserve(1, m.viewDealer()),
		"",
		ui.SectionLabel("you", ""),
		ui.Reserve(4, m.viewPlayers()),
		"",
		ui.BetSelector(m.bet()),
		"",
		ui.Reserve(1, status),
	)

	return ui.Screen("Blackjack", m.bank, m.width, m.height, body, m.hints())
}

func (m *Model) hints() string {
	switch m.phase {
	case phasePlayer:
		h := &m.players[m.active]
		s := "H hit · S stand"
		if len(h.cards) == 2 {
			s += " · D double"
			if m.canSplit(h) {
				s += " · P split"
			}
		}
		return s + " · M menu"
	case phaseResult:
		return "ENTER deal again · M menu"
	default:
		return "‹ › bet · ENTER deal · M menu"
	}
}

func (m *Model) viewDealer() string {
	_, dShown := m.visibleCounts()
	cards := m.dealer
	if dShown < len(cards) {
		cards = cards[:dShown]
	}
	if m.holeDown && len(cards) >= 2 {
		return ui.RenderHand(cards[:1], 1)
	}
	return ui.RenderHand(cards, 0)
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
		switch {
		case h.bust:
			tag = ui.AccentText.Render("BUST " + tag)
		case h.doubled:
			tag = ui.Subtle.Render(tag + " ×2")
		default:
			tag = ui.Subtle.Render(tag)
		}
		col := lipgloss.JoinVertical(lipgloss.Left, row, tag)
		if m.phase == phasePlayer && i == m.active && len(m.players) > 1 {
			col = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).
				BorderForeground(ui.Accent).Padding(0, 1).Render(col)
		}
		cols = append(cols, col)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, joinGap(cols, "   ")...)
}

func joinGap(cols []string, gap string) []string {
	out := make([]string, 0, len(cols)*2)
	for i, c := range cols {
		if i > 0 {
			out = append(out, gap)
		}
		out = append(out, c)
	}
	return out
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
		return "blackjack"
	}
	if soft && v <= 21 {
		return fmt.Sprintf("soft %d", v)
	}
	return fmt.Sprintf("%d", v)
}
