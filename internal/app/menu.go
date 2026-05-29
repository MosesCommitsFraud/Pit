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
	title := ui.AccentText.Render("▌") + ui.Chip.Render("P I T") + ui.AccentText.Render("▐") +
		"   " + ui.Subtle.Render(ui.Caps("terminal casino"))

	var body, hints string
	if m.stage == stageMode {
		body = m.viewMode()
		hints = "↑↓ move · ENTER select · CTRL+C quit"
	} else {
		body = m.viewGames()
		hints = "↑↓ move · ENTER play · ESC change mode · CTRL+C quit"
	}

	w := m.width
	if w < 1 {
		w = 1
	}
	h := m.height - 4
	if h < 1 {
		h = 1
	}
	content := lipgloss.NewStyle().Padding(1, 2).Height(h).Render(body)
	return lipgloss.JoinVertical(lipgloss.Left,
		" "+title,
		ui.Rule(w),
		content,
		ui.Rule(w),
		" "+ui.HelpBar.Render(hints),
	)
}

func (m *menu) viewMode() string {
	modes := []struct{ name, blurb string }{
		{"Sandbox", "Unlimited play money. For the love of the game."},
		{"Career", "Persistent bankroll. Collect 1,000 chips every day."},
	}
	rows := []string{ui.Label.Render("CHOOSE YOUR TABLE"), ""}
	for i, md := range modes {
		rows = append(rows, m.row(i == m.modeIdx, md.name, md.blurb))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *menu) viewGames() string {
	rows := []string{
		ui.SectionLabel(m.bank.Mode().String(), ui.Money(m.bank.Balance())),
	}
	if m.dailyNote != "" {
		rows = append(rows, ui.Subtle.Render(m.dailyNote))
	}
	rows = append(rows, "", ui.Label.Render("SELECT A GAME"), "")
	for i, g := range m.games {
		rows = append(rows, m.row(i == m.gameIdx, g.Title, g.Blurb))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// row renders one selectable list entry, highlighted when focused.
func (m *menu) row(focused bool, name, blurb string) string {
	if focused {
		return ui.Selected.Render("▸ "+ui.Caps(name)) + "  " + ui.Subtle.Render(blurb)
	}
	return ui.Unselected.Render("  "+ui.Caps(name)) + "  " + ui.Subtle.Render(blurb)
}
