package booksun

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"pit/internal/ui"
)

const cellW = 5

var (
	colSep   = lipgloss.NewStyle().Foreground(ui.Faint).Render("  │  ")
	winCell  = lipgloss.NewStyle().Bold(true).Foreground(ui.Black).Background(ui.Accent)
	bookCell = lipgloss.NewStyle().Bold(true).Foreground(ui.Accent)
	highCell = lipgloss.NewStyle().Bold(true).Foreground(ui.Bright)
	lowCell  = lipgloss.NewStyle().Foreground(ui.Text)
)

// renderCell draws one reel symbol as a fixed-width centered glyph, matching
// the Slots machine look.
func (m *Model) renderCell(r, row int) string {
	idx := m.grid[r][row]
	var st lipgloss.Style
	switch {
	case m.winCells[[2]int{r, row}]:
		st = winCell
	case idx == book || idx == symHero:
		st = bookCell
	case idx <= symAce:
		st = lowCell
	default:
		st = highCell
	}
	return st.Width(cellW).Align(lipgloss.Center).Render(syms[idx].glyph)
}

// viewMachine renders the 5x3 grid inside a double-border frame.
func (m *Model) viewMachine() string {
	rows := make([]string, rowsN)
	for row := 0; row < rowsN; row++ {
		cells := make([]string, 0, reels*2)
		for r := 0; r < reels; r++ {
			if r > 0 {
				cells = append(cells, colSep)
			}
			cells = append(cells, m.renderCell(r, row))
		}
		rows[row] = lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	}
	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ui.Faint).
		Padding(1, 2).
		Render(grid)
}

func (m *Model) View() string {
	info := ui.BetSelector(m.stake()) +
		ui.Subtle.Render("   line ") + ui.Heading.Render(ui.Money(m.lineBet())) +
		ui.Subtle.Render(fmt.Sprintf(" · %d lines", numLines))

	free := ""
	if m.inFree {
		free = ui.AccentText.Render(fmt.Sprintf("FREE SPINS  %d", m.freeLeft)) +
			ui.Subtle.Render("   expanding ") +
			ui.AccentText.Render(syms[m.special].glyph+" "+syms[m.special].name)
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
		ui.SectionLabel("reels", ""),
		m.viewMachine(),
		"",
		info,
		ui.Reserve(1, free),
		"",
		ui.Reserve(1, status),
	)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, "      ", m.viewPaytable())
	return ui.Screen("Book of Ra", m.bank, m.width, m.height, body, m.hints())
}

func (m *Model) viewPaytable() string {
	order := []int{symHero, symMask, symScarab, symIdol, symAce, symKing, symQueen, symJack, symTen}
	nameCol := lipgloss.NewStyle().Width(13)
	rows := []string{ui.SectionLabel("pays", "3 / 4 / 5")}
	for _, s := range order {
		p := syms[s].pay
		gs := highCell
		switch {
		case s == symHero:
			gs = bookCell
		case s <= symAce:
			gs = lowCell
		}
		// icon symbols show "<icon> NAME"; royals are just the letter
		label := gs.Render(syms[s].glyph)
		if s > symAce {
			label += " " + ui.Subtle.Render(syms[s].name)
		}
		rows = append(rows, nameCol.Render(label)+
			ui.Subtle.Render(fmt.Sprintf("%4d /%4d /%5d", p[3], p[4], p[5])))
	}
	rows = append(rows, "",
		bookCell.Render(syms[book].glyph)+" "+ui.AccentText.Render("BOOK")+
			ui.Subtle.Render("  wild + scatter"),
		ui.Subtle.Render("3 books = 10 free spins"))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *Model) hints() string {
	if m.inFree {
		return "SPACE free spin · M menu"
	}
	return "SPACE spin · ‹ › bet · M menu"
}
