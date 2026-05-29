package booksun

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"pit/internal/ui"
)

var (
	cellBracket = lipgloss.NewStyle().Foreground(ui.Faint)
	cellNormal  = lipgloss.NewStyle().Foreground(ui.Bright)
	cellRoyal   = lipgloss.NewStyle().Foreground(ui.Text)
	cellAccent  = lipgloss.NewStyle().Bold(true).Foreground(ui.Accent)
	cellWin     = lipgloss.NewStyle().Bold(true).Foreground(ui.Black).Background(ui.Accent)
)

func centerTok(s string, w int) string {
	for len(s) < w {
		if len(s)%2 == 0 {
			s = s + " "
		} else {
			s = " " + s
		}
	}
	return s
}

func (m *Model) renderCell(r, row int) string {
	idx := m.grid[r][row]
	tok := centerTok(syms[idx].token, 4)
	var st lipgloss.Style
	switch {
	case m.winCells[[2]int{r, row}]:
		st = cellWin
	case idx == book || idx == symHero || idx == symMask:
		st = cellAccent
	case idx <= symAce:
		st = cellRoyal
	default:
		st = cellNormal
	}
	return cellBracket.Render("[ ") + st.Render(tok) + cellBracket.Render(" ]")
}

func (m *Model) viewGrid() string {
	rows := make([]string, rowsN)
	for row := 0; row < rowsN; row++ {
		cells := make([]string, 0, reels*2)
		for r := 0; r < reels; r++ {
			if r > 0 {
				cells = append(cells, " ")
			}
			cells = append(cells, m.renderCell(r, row))
		}
		rows[row] = lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *Model) View() string {
	info := ui.Label.Render("BET") + " " + ui.AccentText.Render(ui.Money(m.stake())) +
		ui.Subtle.Render(" ‹ ›   ") +
		ui.Label.Render("LINE") + " " + ui.Heading.Render(ui.Money(m.lineBet())) +
		ui.Subtle.Render(fmt.Sprintf("   %d LINES", numLines))

	free := ""
	if m.inFree {
		free = ui.AccentText.Render(fmt.Sprintf("FREE SPINS  %d", m.freeLeft)) +
			ui.Subtle.Render("   expanding: ") + ui.Heading.Render(syms[m.special].token)
	}

	var status string
	switch {
	case m.phase == phaseResult && m.lastWin > 0:
		status = ui.WinText.Render(strings.ToUpper(m.msg) + "  +" + ui.Money(m.lastWin))
	case m.phase == phaseResult:
		status = ui.LoseText.Render(m.msg)
	default:
		status = ui.Subtle.Render(m.msg)
	}

	left := lipgloss.JoinVertical(lipgloss.Left,
		ui.SectionLabel("book of the sun", ""),
		m.viewGrid(),
		"",
		info,
		ui.Reserve(1, free),
		"",
		ui.Reserve(1, status),
	)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, "      ", m.viewPaytable())
	return ui.Screen("Book of the Sun", m.bank, m.width, m.height, body, m.hints())
}

func (m *Model) viewPaytable() string {
	order := []int{symHero, symMask, symScarab, symIdol, symAce, symKing, symQueen, symJack, symTen}
	rows := []string{ui.SectionLabel("pays", "3 / 4 / 5")}
	for _, s := range order {
		p := syms[s].pay
		rows = append(rows,
			ui.Heading.Render(centerTok(syms[s].token, 4))+ui.Subtle.Render(
				fmt.Sprintf("  %d / %d / %d", p[3], p[4], p[5])))
	}
	rows = append(rows, "",
		ui.AccentText.Render("BOOK")+ui.Subtle.Render("  wild + scatter"),
		ui.Subtle.Render("3 books = 10 free spins"))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *Model) hints() string {
	if m.inFree {
		return "SPACE free spin · M menu"
	}
	return "SPACE spin · ‹ › bet · M menu"
}
