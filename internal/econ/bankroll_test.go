package econ

import "testing"

func TestApply(t *testing.T) {
	b := New(ModeSandbox, 100)
	if got := b.Apply(50); got != 150 {
		t.Fatalf("win: got %d want 150", got)
	}
	if got := b.Apply(-100); got != 50 {
		t.Fatalf("loss: got %d want 50", got)
	}
}

func TestApplyClampsAtZero(t *testing.T) {
	b := New(ModeCareer, 30)
	if got := b.Apply(-100); got != 0 {
		t.Fatalf("bust: got %d want 0 (no negative balance)", got)
	}
}

func TestCanCover(t *testing.T) {
	b := New(ModeSandbox, 100)
	cases := []struct {
		amt  int64
		want bool
	}{
		{50, true}, {100, true}, {101, false}, {0, false}, {-5, false},
	}
	for _, c := range cases {
		if got := b.CanCover(c.amt); got != c.want {
			t.Errorf("CanCover(%d) = %v want %v", c.amt, got, c.want)
		}
	}
}
