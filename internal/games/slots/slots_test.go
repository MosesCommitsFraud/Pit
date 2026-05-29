package slots

import (
	"testing"

	"pit/internal/econ"
	"pit/internal/games"
)

// settleDelta forces a result and returns the net bankroll delta the game emits.
func settleDelta(t *testing.T, betIdx, r0, r1, r2 int) int64 {
	t.Helper()
	m := New(econ.New(econ.ModeSandbox, 100000), 80, 24).(*Model)
	m.betIdx = betIdx
	m.reels[0].result = r0
	m.reels[1].result = r1
	m.reels[2].result = r2
	cmd := m.settle()
	if cmd == nil {
		t.Fatal("settle returned no command")
	}
	msg, ok := cmd().(games.SettleMsg)
	if !ok {
		t.Fatalf("expected SettleMsg, got %T", cmd())
	}
	return msg.Delta
}

func TestSlotsPayouts(t *testing.T) {
	bet := bets[0] // 10

	// three sevens (idx 5, pay3 = 75): return 750, net +740
	if got := settleDelta(t, 0, 5, 5, 5); got != 75*bet-bet {
		t.Errorf("three 7s: delta=%d want %d", got, 75*bet-bet)
	}

	// three cherries (idx 0, pay3 = 3): return 30, net +20
	if got := settleDelta(t, 0, 0, 0, 0); got != 3*bet-bet {
		t.Errorf("three cherries: delta=%d want %d", got, 3*bet-bet)
	}

	// two cherries: return 2*bet, net +bet
	if got := settleDelta(t, 0, 0, 0, 1); got != 2*bet-bet {
		t.Errorf("two cherries: delta=%d want %d", got, 2*bet-bet)
	}

	// no match, no cherries: lose the bet
	if got := settleDelta(t, 0, 1, 2, 3); got != -bet {
		t.Errorf("loss: delta=%d want %d", got, -bet)
	}
}
