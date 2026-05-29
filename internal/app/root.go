// Package app wires the casino together: a root model that owns the bankroll
// and routes between the menu and the active game.
package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"pit/internal/econ"
	"pit/internal/games"
	"pit/internal/store"
)

// Root is the top-level Bubble Tea model.
type Root struct {
	store    *store.Store // nil if career persistence is unavailable
	bank     *econ.Bankroll
	menu     *menu
	scene    tea.Model // active game, or nil when in the menu
	registry []games.Entry
	width    int
	height   int
}

// New builds the root model. store may be nil to disable career persistence.
func New(st *store.Store, reg []games.Entry) *Root {
	return &Root{
		store:    st,
		menu:     newMenu(reg),
		registry: reg,
	}
}

func (r *Root) Init() tea.Cmd { return r.menu.Init() }

func (r *Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width, r.height = msg.Width, msg.Height
		// forward to whichever model is showing
		return r, r.forward(msg)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return r, tea.Quit
		}
		return r, r.forward(msg)

	case chooseModeMsg:
		r.enterMode(msg.mode)
		return r, nil

	case startGameMsg:
		return r.launch(msg.id)

	case games.SettleMsg:
		r.bank.Apply(msg.Delta)
		if r.bank.Mode() == econ.ModeCareer && r.store != nil {
			_ = r.store.SaveBalance(r.bank.Balance())
		}
		// fall through so the game can also react to its own settle
		return r, r.forward(msg)

	case games.QuitToMenuMsg:
		r.scene = nil
		return r, nil

	default:
		return r, r.forward(msg)
	}
}

// forward sends a message to the active model (game if any, else the menu) and
// stores the updated model.
func (r *Root) forward(msg tea.Msg) tea.Cmd {
	if r.scene != nil {
		m, cmd := r.scene.Update(msg)
		r.scene = m
		return cmd
	}
	m, cmd := r.menu.Update(msg)
	r.menu = m.(*menu)
	return cmd
}

// enterMode sets up the bankroll for the chosen mode and moves the menu to the
// game list. In career mode it loads the persisted balance and claims the daily
// stipend.
func (r *Root) enterMode(mode econ.Mode) {
	if mode == econ.ModeCareer && r.store != nil {
		today := time.Now().Format("2006-01-02")
		bal, claimed, err := r.store.ClaimDaily(today, econ.DailyStipend)
		if err != nil {
			// fall back to whatever load gives us
			st, _ := r.store.Load()
			bal = st.Balance
		}
		r.bank = econ.New(econ.ModeCareer, bal)
		if claimed {
			r.menu.dailyNote = "Daily stipend collected: +1,000 chips"
		} else {
			r.menu.dailyNote = "Daily stipend already collected — come back tomorrow"
		}
	} else {
		r.bank = econ.New(econ.ModeSandbox, econ.SandboxStart)
		r.menu.dailyNote = ""
	}
	r.menu.bank = r.bank
	r.menu.stage = stageGames
	r.menu.gameIdx = 0
}

// launch constructs and enters the game with the given id.
func (r *Root) launch(id string) (tea.Model, tea.Cmd) {
	for _, e := range r.registry {
		if e.ID == id {
			r.scene = e.New(r.bank, r.width, r.height)
			return r, r.scene.Init()
		}
	}
	return r, nil
}

func (r *Root) View() string {
	if r.scene != nil {
		return r.scene.View()
	}
	return r.menu.View()
}
