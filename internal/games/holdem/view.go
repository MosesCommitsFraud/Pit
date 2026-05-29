package holdem

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"pit/internal/ui"
)

func (m *Model) View() string {
	body := lipgloss.JoinVertical(lipgloss.Center,
		ui.Reserve(5, m.viewOpponents()),
		"",
		ui.Reserve(3, m.viewBoard()),
		"",
		ui.Reserve(3, m.viewHuman()),
		"",
		ui.Reserve(2, m.viewStatus()),
	)
	return ui.Screen("Texas Hold'em", m.bank, m.width, m.height, body, m.hints())
}

func (m *Model) viewOpponents() string {
	cols := make([]string, 0, len(m.seats)-1)
	reveal := m.phase == phaseShowdown
	for i := 1; i < len(m.seats); i++ {
		s := &m.seats[i]
		var cards string
		switch {
		case s.folded:
			cards = ui.Subtle.Render("  —    —  ")
		case reveal:
			cards = ui.RenderHand(s.hole, 0)
		default:
			cards = ui.RenderHand(nil, 2)
		}
		name := s.name
		if i == m.button {
			name += " (BTN)"
		}
		nameStyle := ui.Heading
		if i == m.toAct && (m.phase == phaseBot || m.phase == phaseHuman) {
			nameStyle = ui.AccentText
		}
		if s.folded {
			nameStyle = ui.Subtle
		}
		info := nameStyle.Render(name) + "\n" +
			ui.Subtle.Render(ui.Money(s.stack)) + "\n" +
			actionTag(s)
		col := lipgloss.JoinVertical(lipgloss.Center, cards, info)
		cols = append(cols, lipgloss.NewStyle().Padding(0, 2).Render(col))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

func actionTag(s *seat) string {
	if s.street > 0 && !s.folded {
		return ui.AccentText.Render(ui.Money(s.street))
	}
	if s.lastAct != "" {
		return ui.Subtle.Render(s.lastAct)
	}
	return " "
}

func (m *Model) viewBoard() string {
	slots := make([]string, 0, 9)
	for i, c := range m.board {
		if i > 0 {
			slots = append(slots, " ")
		}
		slots = append(slots, ui.RenderCard(c))
	}
	for n := len(m.board); n < 5; n++ {
		if n > 0 {
			slots = append(slots, " ")
		}
		slots = append(slots, ui.RenderEmpty())
	}
	board := lipgloss.JoinHorizontal(lipgloss.Top, slots...)
	pot := ui.Label.Render("POT") + "  " + ui.AccentText.Render(ui.Money(m.pot+m.tableStreet()))
	return lipgloss.JoinVertical(lipgloss.Center, board, pot)
}

func (m *Model) viewHuman() string {
	s := m.human()
	cards := ui.Subtle.Render("(waiting)")
	if len(s.hole) > 0 {
		cards = ui.RenderHand(s.hole, 0)
	}
	name := "YOU"
	if m.button == 0 {
		name += " (BTN)"
	}
	nameStyle := ui.Heading
	if m.toAct == 0 && m.phase == phaseHuman {
		nameStyle = ui.AccentText
	}
	if s.folded {
		name += " · folded"
		nameStyle = ui.Subtle
	}
	info := nameStyle.Render(name) + "   " + ui.Subtle.Render("stack "+ui.Money(s.stack)) + "   " + actionTag(s)
	return lipgloss.JoinVertical(lipgloss.Center, cards, info)
}

func (m *Model) viewStatus() string {
	switch m.phase {
	case phaseIdle:
		msg := "press ENTER to deal"
		if m.result != "" {
			msg = m.result
		}
		return ui.Subtle.Render(msg)
	case phaseBot:
		return ui.Subtle.Render(m.seats[m.toAct].name + " is thinking…")
	case phaseRunout:
		return ui.Subtle.Render("running it out…")
	case phaseShowdown:
		banner := ui.Banner(ui.ResultPush, m.result, 0)
		if m.delta > 0 {
			banner = ui.Banner(ui.ResultWin, m.result, m.delta)
		} else if m.delta < 0 {
			banner = ui.Banner(ui.ResultLose, m.result, m.delta)
		}
		return banner + "\n" + ui.Subtle.Render("ENTER for next hand")
	case phaseHuman:
		return m.viewActions()
	}
	return ""
}

func (m *Model) viewActions() string {
	call := m.callAmount(0)
	callLabel := "check"
	if call > 0 {
		callLabel = fmt.Sprintf("call %s", ui.Money(call))
	}
	parts := ui.AccentText.Render("YOUR TURN") + "    " +
		ui.Unselected.Render("[F] fold") + " " +
		ui.Unselected.Render("[C] "+callLabel)
	if len(m.raiseSizes) > 0 {
		verb := "raise to"
		if m.currentBet == 0 {
			verb = "bet"
		}
		size := m.raiseSizes[m.raiseIdx]
		raise := ui.AccentText.Render(fmt.Sprintf("[R] %s %s", verb, ui.Money(size)))
		parts += " " + raise + ui.Subtle.Render(" ‹ ›")
	}
	return parts
}

func (m *Model) hints() string {
	switch m.phase {
	case phaseHuman:
		return "F fold · C check/call · ‹ › size · R raise · M menu"
	case phaseIdle, phaseShowdown:
		return "ENTER deal · M menu"
	default:
		return "M menu"
	}
}
