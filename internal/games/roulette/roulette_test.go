package roulette

import "testing"

func targetByKind(k betKind) target {
	for _, t := range targets {
		if t.kind == k {
			return t
		}
	}
	panic("no such target")
}

func TestWins(t *testing.T) {
	cases := []struct {
		kind   betKind
		number int
		spun   int
		want   bool
	}{
		{betRed, 0, 1, true},  // 1 is red
		{betRed, 0, 2, false}, // 2 is black
		{betBlack, 0, 2, true},
		{betEven, 0, 0, false}, // zero is never even/odd
		{betOdd, 0, 0, false},
		{betEven, 0, 4, true},
		{betLow, 0, 18, true},
		{betHigh, 0, 18, false},
		{betDozen1, 0, 12, true},
		{betDozen2, 0, 13, true},
		{betDozen3, 0, 36, true},
		{betStraight, 17, 17, true},
		{betStraight, 17, 18, false},
	}
	for _, c := range cases {
		if got := targetByKind(c.kind).wins(c.spun, c.number); got != c.want {
			t.Errorf("kind=%d num=%d spun=%d: got %v want %v", c.kind, c.number, c.spun, got, c.want)
		}
	}
}

func TestSettlePayouts(t *testing.T) {
	// straight win pays 35:1, others lose
	m := New(nil, 80, 24).(*Model)
	m.bank = nil
	// place 10 on straight (index of straight) and 10 on red
	for i, tg := range targets {
		if tg.kind == betStraight {
			m.wagers[i] = 10
		}
		if tg.kind == betRed {
			m.wagers[i] = 10
		}
	}
	m.number = 17
	m.result = 17 // 17 is black → red loses, straight wins
	m.settle()
	// straight: +350, red: -10 → net +340
	if m.delta != 340 {
		t.Errorf("net delta = %d want 340", m.delta)
	}
}
