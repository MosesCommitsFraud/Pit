package holdem

import (
	"testing"

	"pit/internal/deck"
	"pit/internal/econ"
	"pit/internal/games"
)

func cd(r deck.Rank, s deck.Suit) deck.Card { return deck.Card{Rank: r, Suit: s} }

// A flush must rank stronger (lower) than a straight on the same board.
func TestFlushBeatsStraight(t *testing.T) {
	board := []deck.Card{
		cd(deck.Two, deck.Hearts),
		cd(deck.Seven, deck.Hearts),
		cd(deck.Nine, deck.Hearts),
		cd(deck.Three, deck.Clubs),
		cd(deck.Four, deck.Diamonds),
	}
	flush := []deck.Card{cd(deck.Ace, deck.Hearts), cd(deck.King, deck.Hearts)}
	straight := []deck.Card{cd(deck.Five, deck.Clubs), cd(deck.Six, deck.Diamonds)}

	rf := rankOf(flush, board)
	rs := rankOf(straight, board)
	if rf >= rs {
		t.Fatalf("flush rank %d should be < straight rank %d", rf, rs)
	}
	if strengthFromRank(rf) <= strengthFromRank(rs) {
		t.Fatalf("flush strength should exceed straight strength")
	}
}

// Showdown should award the whole pot to the best hand and settle the human's
// net correctly.
func TestShowdownAwardsBestHand(t *testing.T) {
	m := New(econ.New(econ.ModeSandbox, 1000), 80, 24).(*Model)
	board := []deck.Card{
		cd(deck.Two, deck.Hearts),
		cd(deck.Seven, deck.Hearts),
		cd(deck.Nine, deck.Hearts),
		cd(deck.Three, deck.Clubs),
		cd(deck.Four, deck.Diamonds),
	}
	m.board = board
	m.startBank = 100
	m.pot = 200

	// You: flush; Ada: straight; Boyd & Cleo folded.
	m.seats[0].hole = []deck.Card{cd(deck.Ace, deck.Hearts), cd(deck.King, deck.Hearts)}
	m.seats[0].stack = 0 // contributed 100 (startBank 100 -> 0)
	m.seats[1].hole = []deck.Card{cd(deck.Five, deck.Clubs), cd(deck.Six, deck.Diamonds)}
	m.seats[2].folded = true
	m.seats[3].folded = true

	cmd := m.showdown()
	msg, ok := cmd().(games.SettleMsg)
	if !ok {
		t.Fatalf("expected SettleMsg got %T", cmd())
	}
	if m.seats[0].stack != 200 {
		t.Errorf("winner stack = %d want 200", m.seats[0].stack)
	}
	// contributed 100, won 200 pot -> net +100
	if msg.Delta != 100 {
		t.Errorf("delta = %d want 100", msg.Delta)
	}
}

// A losing showdown should produce a negative settle.
func TestShowdownLoss(t *testing.T) {
	m := New(econ.New(econ.ModeSandbox, 1000), 80, 24).(*Model)
	board := []deck.Card{
		cd(deck.Two, deck.Hearts),
		cd(deck.Seven, deck.Hearts),
		cd(deck.Nine, deck.Hearts),
		cd(deck.Three, deck.Clubs),
		cd(deck.Four, deck.Diamonds),
	}
	m.board = board
	m.startBank = 100
	m.pot = 200
	m.seats[0].hole = []deck.Card{cd(deck.Five, deck.Clubs), cd(deck.Six, deck.Diamonds)} // straight
	m.seats[0].stack = 0
	m.seats[1].hole = []deck.Card{cd(deck.Ace, deck.Hearts), cd(deck.King, deck.Hearts)} // flush
	m.seats[2].folded = true
	m.seats[3].folded = true

	cmd := m.showdown()
	msg := cmd().(games.SettleMsg)
	if msg.Delta != -100 {
		t.Errorf("delta = %d want -100 (lost the 100 contributed)", msg.Delta)
	}
}
