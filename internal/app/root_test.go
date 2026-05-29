package app

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"pit/internal/econ"
	"pit/internal/games"
	"pit/internal/store"
)

func tempStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "save.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// drive feeds a message through the root and returns it (cmds ignored).
func drive(r *Root, msg tea.Msg) {
	m, _ := r.Update(msg)
	*r = *m.(*Root)
}

func TestCareerPersistsAcrossSessions(t *testing.T) {
	st := tempStore(t)

	// session 1: enter career (fresh DB → daily stipend of 1000)
	r1 := New(st, Registry())
	drive(r1, tea.WindowSizeMsg{Width: 80, Height: 24})
	drive(r1, chooseModeMsg{mode: econ.ModeCareer})
	if got := r1.bank.Balance(); got != econ.DailyStipend {
		t.Fatalf("after first daily claim: balance %d want %d", got, econ.DailyStipend)
	}

	// win 500 → applied and persisted
	drive(r1, games.SettleMsg{Delta: 500})
	if got := r1.bank.Balance(); got != 1500 {
		t.Fatalf("after win: balance %d want 1500", got)
	}
	if s, _ := st.Load(); s.Balance != 1500 {
		t.Fatalf("persisted balance %d want 1500", s.Balance)
	}

	// session 2 (same DB, same day): no second stipend, balance retained
	r2 := New(st, Registry())
	drive(r2, tea.WindowSizeMsg{Width: 80, Height: 24})
	drive(r2, chooseModeMsg{mode: econ.ModeCareer})
	if got := r2.bank.Balance(); got != 1500 {
		t.Fatalf("relaunch: balance %d want 1500 (winnings kept, no double stipend)", got)
	}
}

func TestSandboxNeverPersists(t *testing.T) {
	st := tempStore(t)
	r := New(st, Registry())
	drive(r, tea.WindowSizeMsg{Width: 80, Height: 24})
	drive(r, chooseModeMsg{mode: econ.ModeSandbox})
	if got := r.bank.Balance(); got != econ.SandboxStart {
		t.Fatalf("sandbox start %d want %d", got, econ.SandboxStart)
	}
	drive(r, games.SettleMsg{Delta: -2000})
	// store must be untouched (still the initial zero row)
	if s, _ := st.Load(); s.Balance != 0 {
		t.Fatalf("sandbox leaked to store: balance %d want 0", s.Balance)
	}
}

func TestLaunchAndQuitGame(t *testing.T) {
	r := New(nil, Registry())
	drive(r, tea.WindowSizeMsg{Width: 100, Height: 30})
	drive(r, chooseModeMsg{mode: econ.ModeSandbox})
	for _, e := range Registry() {
		drive(r, startGameMsg{id: e.ID})
		if r.scene == nil {
			t.Fatalf("game %q did not launch", e.ID)
		}
		// every game must render without panicking
		if r.View() == "" {
			t.Fatalf("game %q rendered empty", e.ID)
		}
		drive(r, games.QuitToMenuMsg{})
		if r.scene != nil {
			t.Fatalf("game %q did not return to menu", e.ID)
		}
	}
}
