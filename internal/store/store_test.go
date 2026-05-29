package store

import (
	"path/filepath"
	"testing"
)

func openTemp(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "save.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestClaimDailyGrantsOncePerDay(t *testing.T) {
	s := openTemp(t)

	bal, claimed, err := s.ClaimDaily("2026-05-29", 1000)
	if err != nil || !claimed || bal != 1000 {
		t.Fatalf("first claim: bal=%d claimed=%v err=%v want 1000,true,nil", bal, claimed, err)
	}

	// same day again: no grant
	bal, claimed, err = s.ClaimDaily("2026-05-29", 1000)
	if err != nil || claimed || bal != 1000 {
		t.Fatalf("repeat claim: bal=%d claimed=%v want 1000,false", bal, claimed)
	}

	// next day: grant again on top of kept balance
	bal, claimed, err = s.ClaimDaily("2026-05-30", 1000)
	if err != nil || !claimed || bal != 2000 {
		t.Fatalf("next day: bal=%d claimed=%v want 2000,true", bal, claimed)
	}
}

func TestSaveBalanceRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.db")

	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.ClaimDaily("2026-05-29", 1000); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveBalance(4250); err != nil {
		t.Fatal(err)
	}
	s.Close()

	// reopen: balance and claim date persist
	s2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	st, err := s2.Load()
	if err != nil {
		t.Fatal(err)
	}
	if st.Balance != 4250 {
		t.Errorf("balance = %d want 4250", st.Balance)
	}
	if st.LastClaim != "2026-05-29" {
		t.Errorf("last claim = %q want 2026-05-29", st.LastClaim)
	}
	// already claimed today → still no double grant after reopen
	if _, claimed, _ := s2.ClaimDaily("2026-05-29", 1000); claimed {
		t.Error("claimed again after reopen on same day")
	}
}
