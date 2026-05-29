package holdem

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"pit/internal/deck"
	"pit/internal/ui"
)

func (m *Model) View() string {
	header := ui.Header("Texas Hold'em", m.bank, m.width)

	opp := m.viewOpponents()
	board := m.viewBoard()
	me := m.viewHuman()

	body := lipgloss.JoinVertical(lipgloss.Center,
		ui.Reserve(6, opp),
		"",
		ui.Reserve(4, board),
		"",
		ui.Reserve(4, me),
		"",
		ui.Reserve(2, m.viewStatus()),
	)

	center := ui.Stage(m.width, m.height, 78, 20, body)
	return lipgloss.JoinVertical(lipgloss.Left, header, center, m.helpLine())
}

func (m *Model) viewOpponents() string {
	cols := make([]string, 0, len(m.seats)-1)
	reveal := m.phase == phaseShowdown
	for i := 1; i < len(m.seats); i++ {
		s := &m.seats[i]
		var cards string
		if s.folded {
			cards = ui.Subtle.Render("  ╳  ╳ ")
		} else if reveal {
			cards = ui.RenderHand(s.hole, 0)
		} else {
			cards = ui.RenderHand(nil, 2)
		}
		name := s.name
		if i == m.button {
			name += " Ⓑ"
		}
		nameStyle := ui.Heading
		if i == m.toAct && (m.phase == phaseBot || m.phase == phaseHuman) {
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(ui.Gold)
		}
		if s.folded {
			nameStyle = ui.Subtle
		}
		info := nameStyle.Render(name) + "\n" +
			ui.Subtle.Render(ui.Money(s.stack)) + "\n" +
			actionTag(s)
		col := lipgloss.JoinVertical(lipgloss.Center, cards, info)
		cols = append(cols, lipgloss.NewStyle().Padding(0, 1).Render(col))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

func actionTag(s *seat) string {
	if s.street > 0 && !s.folded {
		return ui.WinText.Render(ui.Money(s.street))
	}
	if s.lastAct != "" {
		return ui.Subtle.Render(s.lastAct)
	}
	return " "
}

func (m *Model) viewBoard() string {
	slots := make([]string, 0, 5)
	for _, c := range m.board {
		slots = append(slots, ui.RenderHand([]deck.Card{c}, 0))
	}
	placeholder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(ui.Border).
		Render("     \n     \n     ")
	for len(slots) < 5 {
		slots = append(slots, placeholder)
	}
	board := lipgloss.JoinHorizontal(lipgloss.Top, slots...)
	potLine := ui.Heading.Render("Pot: ") +
		lipgloss.NewStyle().Bold(true).Foreground(ui.Gold).Render(ui.Money(m.pot+m.tableStreet()))
	return lipgloss.JoinVertical(lipgloss.Center, board, potLine)
}

func (m *Model) viewHuman() string {
	s := m.human()
	var cards string
	if len(s.hole) > 0 {
		cards = ui.RenderHand(s.hole, 0)
	} else {
		cards = ui.Subtle.Render("(waiting)")
	}
	name := "You"
	if m.button == 0 {
		name += " Ⓑ"
	}
	nameStyle := ui.Heading
	if m.toAct == 0 && m.phase == phaseHuman {
		nameStyle = lipgloss.NewStyle().Bold(true).Foreground(ui.Gold)
	}
	if s.folded {
		name += "  (folded)"
		nameStyle = ui.Subtle
	}
	info := nameStyle.Render(name) + "   " + ui.Subtle.Render("stack "+ui.Money(s.stack)) + "   " + actionTag(s)
	return lipgloss.JoinVertical(lipgloss.Center, cards, info)
}

func (m *Model) viewStatus() string {
	switch m.phase {
	case phaseIdle:
		msg := "Press ENTER to deal"
		if m.result != "" {
			msg = m.result
		}
		return ui.Subtle.Render(msg)
	case phaseBot:
		return ui.Subtle.Render(m.seats[m.toAct].name + " is thinking…")
	case phaseRunout:
		return ui.Subtle.Render("Running it out…")
	case phaseShowdown:
		banner := ui.Banner(ui.ResultPush, m.result, 0)
		if m.delta > 0 {
			banner = ui.Banner(ui.ResultWin, m.result, m.delta)
		} else if m.delta < 0 {
			banner = ui.Banner(ui.ResultLose, m.result, m.delta)
		}
		return banner + "\n" + ui.Subtle.Render("Press ENTER for next hand")
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
	parts := ui.Heading.Render("Your turn:  ") +
		ui.Unselected.Render("[f] fold") + "  " +
		ui.Unselected.Render("[c] "+callLabel)
	if len(m.raiseSizes) > 0 {
		verb := "raise to"
		if m.currentBet == 0 {
			verb = "bet"
		}
		size := m.raiseSizes[m.raiseIdx]
		raise := lipgloss.NewStyle().Bold(true).Foreground(ui.Gold).
			Render(fmt.Sprintf("[r] %s %s", verb, ui.Money(size)))
		parts += "  " + raise + ui.Subtle.Render("  (←/→ size)")
	}
	return parts
}

func (m *Model) helpLine() string {
	switch m.phase {
	case phaseHuman:
		return ui.Help("f fold · c check/call · ←/→ size · r raise · m menu")
	case phaseIdle, phaseShowdown:
		return ui.Help("enter deal · m menu")
	default:
		return ui.Help("m menu")
	}
}
