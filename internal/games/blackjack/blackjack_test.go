package blackjack

import (
	"testing"

	"pit/internal/deck"
	"pit/internal/econ"
	"pit/internal/games"
)

func c(r deck.Rank, s deck.Suit) deck.Card { return deck.Card{Rank: r, Suit: s} }

func TestHandValue(t *testing.T) {
	cases := []struct {
		name  string
		cards []deck.Card
		total int
		soft  bool
	}{
		{"natural", []deck.Card{c(deck.Ace, 0), c(deck.King, 1)}, 21, true},
		{"ace reduced", []deck.Card{c(deck.Ace, 0), c(deck.Nine, 1), c(deck.King, 2)}, 20, false},
		{"two aces", []deck.Card{c(deck.Ace, 0), c(deck.Ace, 1), c(deck.Nine, 2)}, 21, true},
		{"hard 20", []deck.Card{c(deck.King, 0), c(deck.Queen, 1)}, 20, false},
		{"bust", []deck.Card{c(deck.King, 0), c(deck.Queen, 1), c(deck.Five, 2)}, 25, false},
	}
	for _, tc := range cases {
		got, soft := handValue(tc.cards)
		if got != tc.total || soft != tc.soft {
			t.Errorf("%s: got (%d,%v) want (%d,%v)", tc.name, got, soft, tc.total, tc.soft)
		}
	}
}

// resolveDelta forces hands and returns the emitted net delta.
func resolveDelta(t *testing.T, bet int64, wasSplit bool, dealer []deck.Card, players ...[]deck.Card) int64 {
	t.Helper()
	m := New(econ.New(econ.ModeSandbox, 100000), 80, 24).(*Model)
	m.wasSplit = wasSplit
	m.dealer = dealer
	m.players = nil
	for _, pc := range players {
		bv := bet
		// detect doubled by marker not needed here
		m.players = append(m.players, hand{cards: pc, bet: bv, bust: func() bool { v, _ := handValue(pc); return v > 21 }()})
	}
	cmd := m.resolve()
	msg, ok := cmd().(games.SettleMsg)
	if !ok {
		t.Fatalf("expected SettleMsg got %T", cmd())
	}
	return msg.Delta
}

func TestResolveOutcomes(t *testing.T) {
	bet := int64(100)
	K, Q, T, N, A, F := deck.King, deck.Queen, deck.Ten, deck.Nine, deck.Ace, deck.Five

	// player 20 beats dealer 18
	if got := resolveDelta(t, bet, false, []deck.Card{c(K, 0), c(N, 1)}, []deck.Card{c(K, 2), c(T, 3)}); got != bet {
		t.Errorf("win: got %d want %d", got, bet)
	}
	// player bust loses regardless of dealer
	if got := resolveDelta(t, bet, false, []deck.Card{c(K, 0), c(N, 1)}, []deck.Card{c(K, 2), c(Q, 3), c(F, 0)}); got != -bet {
		t.Errorf("bust: got %d want %d", got, -bet)
	}
	// natural blackjack pays 3:2
	if got := resolveDelta(t, bet, false, []deck.Card{c(K, 0), c(N, 1)}, []deck.Card{c(A, 2), c(K, 3)}); got != bet*3/2 {
		t.Errorf("blackjack: got %d want %d", got, bet*3/2)
	}
	// push
	if got := resolveDelta(t, bet, false, []deck.Card{c(K, 0), c(T, 1)}, []deck.Card{c(Q, 2), c(T, 3)}); got != 0 {
		t.Errorf("push: got %d want 0", got)
	}
	// dealer busts, player stands on 18
	if got := resolveDelta(t, bet, false, []deck.Card{c(K, 0), c(F, 1), c(N, 2)}, []deck.Card{c(K, 2), c(deck.Eight, 3)}); got != bet {
		t.Errorf("dealer bust: got %d want %d", got, bet)
	}
}
