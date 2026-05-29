// Package econ models the player's money: the bankroll and the two play modes.
package econ

// Mode is how money behaves for the current session.
type Mode int

const (
	// ModeSandbox is non-persistent play money that resets each launch.
	ModeSandbox Mode = iota
	// ModeCareer is a persistent bankroll topped up by a daily stipend.
	ModeCareer
)

func (m Mode) String() string {
	if m == ModeCareer {
		return "Career"
	}
	return "Sandbox"
}

const (
	// SandboxStart is the play-money balance handed out in sandbox mode.
	SandboxStart int64 = 10_000
	// DailyStipend is the chips granted once per local day in career mode.
	DailyStipend int64 = 1_000
)

// Bankroll is the single source of truth for the player's balance during a
// session. Games read it for display and emit settle deltas; the root model
// is responsible for applying those deltas and persisting in career mode.
type Bankroll struct {
	balance int64
	mode    Mode
}

// New returns a bankroll for the given mode and starting balance.
func New(mode Mode, balance int64) *Bankroll {
	return &Bankroll{balance: balance, mode: mode}
}

// Balance returns the current balance in chips.
func (b *Bankroll) Balance() int64 { return b.balance }

// Mode returns the active play mode.
func (b *Bankroll) Mode() Mode { return b.mode }

// CanCover reports whether the balance can fund a bet of amt.
func (b *Bankroll) CanCover(amt int64) bool { return amt > 0 && amt <= b.balance }

// Apply adjusts the balance by delta (negative for a net loss, positive for a
// net win) and returns the new balance. The balance never goes below zero.
func (b *Bankroll) Apply(delta int64) int64 {
	b.balance += delta
	if b.balance < 0 {
		b.balance = 0
	}
	return b.balance
}
