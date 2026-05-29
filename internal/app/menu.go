package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pit/internal/econ"
	"pit/internal/games"
	"pit/internal/ui"
)

type menuStage int

const (
	stageMode menuStage = iota
	stageGames
)

// menu is the title screen: pick a mode, then pick a game.
type menu struct {
	stage   menuStage
	modeIdx int
	gameIdx int
	games   []games.Entry

	bank      *econ.Bankroll // set once a mode is chosen
	dailyNote string         // career daily-stipend banner

	width, height int
}

func newMenu(reg []games.Entry) *menu {
	return &menu{games: reg}
}

func (m *menu) Init() tea.Cmd { return nil }

func (m *menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch m.stage {
		case stageMode:
			return m.updateMode(msg)
		case stageGames:
			return m.updateGames(msg)
		}
	}
	return m, nil
}

func (m *menu) updateMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.modeIdx > 0 {
			m.modeIdx--
		}
	case "down", "j":
		if m.modeIdx < 1 {
			m.modeIdx++
		}
	case "enter":
		mode := econ.ModeSandbox
		if m.modeIdx == 1 {
			mode = econ.ModeCareer
		}
		return m, func() tea.Msg { return chooseModeMsg{mode: mode} }
	}
	return m, nil
}

func (m *menu) updateGames(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.gameIdx > 0 {
			m.gameIdx--
		}
	case "down", "j":
		if m.gameIdx < len(m.games)-1 {
			m.gameIdx++
		}
	case "esc", "backspace":
		m.stage = stageMode
		m.bank = nil
	case "enter":
		if len(m.games) == 0 {
			return m, nil
		}
		id := m.games[m.gameIdx].ID
		return m, func() tea.Msg { return startGameMsg{id: id} }
	}
	return m, nil
}

func (m *menu) View() string {
	logo := ui.Title.Render(banner)
	var body string
	if m.stage == stageMode {
		body = m.viewMode()
	} else {
		body = m.viewGames()
	}
	content := lipgloss.JoinVertical(lipgloss.Left, logo, "", body)
	return lipgloss.Place(max(m.width, 1), max(m.height, 1),
		lipgloss.Center, lipgloss.Center, ui.Panel.Render(content))
}

func (m *menu) viewMode() string {
	modes := []struct{ name, blurb string }{
		{"Sandbox", "Unlimited play money. For the love of the game."},
		{"Career", "Persistent bankroll. Collect 1,000 chips every day."},
	}
	rows := []string{ui.Heading.Render("Choose your table"), ""}
	for i, md := range modes {
		label := md.name
		if i == m.modeIdx {
			rows = append(rows, ui.Selected.Render("› "+label)+"  "+ui.Subtle.Render(md.blurb))
		} else {
			rows = append(rows, ui.Unselected.Render("  "+label)+"  "+ui.Subtle.Render(md.blurb))
		}
	}
	rows = append(rows, "", ui.Help("↑/↓ move · enter select · ctrl+c quit"))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *menu) viewGames() string {
	rows := []string{}
	head := ui.Heading.Render(m.bank.Mode().String()+" · ") +
		lipgloss.NewStyle().Bold(true).Foreground(ui.Gold).Render(ui.Money(m.bank.Balance()))
	rows = append(rows, head)
	if m.dailyNote != "" {
		rows = append(rows, ui.WinText.Render(m.dailyNote))
	}
	rows = append(rows, "", ui.Heading.Render("Pick a game"), "")
	for i, g := range m.games {
		if i == m.gameIdx {
			rows = append(rows, ui.Selected.Render("› "+g.Title)+"  "+ui.Subtle.Render(g.Blurb))
		} else {
			rows = append(rows, ui.Unselected.Render("  "+g.Title)+"  "+ui.Subtle.Render(g.Blurb))
		}
	}
	rows = append(rows, "", ui.Help("↑/↓ move · enter play · esc change mode · ctrl+c quit"))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

const banner = "  ♠ ♥ P I T ♦ ♣\n  terminal casino"

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
