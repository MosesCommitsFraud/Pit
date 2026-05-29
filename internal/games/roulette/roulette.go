// Package roulette implements European single-zero roulette with an animated
// spinning ball and a board of inside/outside bets.
package roulette

import (
	"math/rand/v2"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"pit/internal/econ"
	"pit/internal/games"
)

// pockets in physical wheel order (European single zero).
var wheel = []int{
	0, 32, 15, 19, 4, 21, 2, 25, 17, 34, 6, 27, 13, 36, 11, 30, 8, 23, 10,
	5, 24, 16, 33, 1, 20, 14, 31, 9, 22, 18, 29, 7, 28, 12, 35, 3, 26,
}

var redSet = map[int]bool{
	1: true, 3: true, 5: true, 7: true, 9: true, 12: true, 14: true, 16: true,
	18: true, 19: true, 21: true, 23: true, 25: true, 27: true, 30: true,
	32: true, 34: true, 36: true,
}

func isRed(n int) bool   { return redSet[n] }
func isBlack(n int) bool { return n != 0 && !redSet[n] }

type betKind int

const (
	betRed betKind = iota
	betBlack
	betEven
	betOdd
	betLow
	betHigh
	betDozen1
	betDozen2
	betDozen3
	betStraight
)

type target struct {
	kind   betKind
	label  string
	payout int64 // profit multiplier per unit staked
}

var targets = []target{
	{betRed, "Red", 1},
	{betBlack, "Black", 1},
	{betEven, "Even", 1},
	{betOdd, "Odd", 1},
	{betLow, "Low 1-18", 1},
	{betHigh, "High 19-36", 1},
	{betDozen1, "1st 12", 2},
	{betDozen2, "2nd 12", 2},
	{betDozen3, "3rd 12", 2},
	{betStraight, "Straight", 35},
}

// wins reports whether a bet on this target (with number for straight) wins for
// the spun pocket n.
func (t target) wins(n, number int) bool {
	switch t.kind {
	case betRed:
		return isRed(n)
	case betBlack:
		return isBlack(n)
	case betEven:
		return n != 0 && n%2 == 0
	case betOdd:
		return n%2 == 1
	case betLow:
		return n >= 1 && n <= 18
	case betHigh:
		return n >= 19 && n <= 36
	case betDozen1:
		return n >= 1 && n <= 12
	case betDozen2:
		return n >= 13 && n <= 24
	case betDozen3:
		return n >= 25 && n <= 36
	case betStraight:
		return n == number
	}
	return false
}

var chips = []int64{5, 10, 25, 100}

type phase int

const (
	phaseBetting phase = iota
	phaseSpin
	phaseResult
)

type tickMsg time.Time

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Model is the roulette game.
type Model struct {
	bank    *econ.Bankroll
	cursor  int     // selected target
	chipIdx int     // selected chip denomination
	number  int     // straight-bet number 0..36
	wagers  []int64 // chips placed on each target index

	wheelPos int // index into wheel during animation
	result   int // final pocket
	frame    int
	stopAt   int
	history  []int

	phase   phase
	delta   int64
	outcome string

	width, height int
}

// New builds a roulette model bound to the bankroll.
func New(b *econ.Bankroll, width, height int) tea.Model {
	return &Model{
		bank:   b,
		wagers: make([]int64, len(targets)),
		width:  width,
		height: height,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) chip() int64 { return chips[m.chipIdx] }

func (m *Model) totalWagered() int64 {
	var t int64
	for _, w := range m.wagers {
		t += w
	}
	return t
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tickMsg:
		if m.phase == phaseSpin {
			return m, m.advance()
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "m", "esc", "q":
			return m, games.QuitToMenu
		}
		if m.phase == phaseSpin {
			return m, nil
		}
		return m.updateBetting(msg)
	}
	return m, nil
}

func (m *Model) updateBetting(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(targets)-1 {
			m.cursor++
		}
	case "left", "h":
		if m.chipIdx > 0 {
			m.chipIdx--
		}
	case "right", "l":
		if m.chipIdx < len(chips)-1 {
			m.chipIdx++
		}
	case "-", "_":
		if targets[m.cursor].kind == betStraight && m.number > 0 {
			m.number--
		}
	case "+", "=":
		if targets[m.cursor].kind == betStraight && m.number < 36 {
			m.number++
		}
	case "enter":
		if m.bank.CanCover(m.totalWagered() + m.chip()) {
			m.wagers[m.cursor] += m.chip()
			m.outcome = ""
		} else {
			m.outcome = "Not enough chips"
		}
	case "backspace", "c":
		for i := range m.wagers {
			m.wagers[i] = 0
		}
		m.outcome = ""
	case " ":
		if m.totalWagered() > 0 {
			return m, m.spin()
		}
		m.outcome = "Place a bet first"
	}
	return m, nil
}

func (m *Model) spin() tea.Cmd {
	m.result = rand.IntN(37)
	m.phase = phaseSpin
	m.frame = 0
	m.stopAt = 38
	m.outcome = ""
	return tick(45 * time.Millisecond)
}

func (m *Model) advance() tea.Cmd {
	m.frame++
	if m.frame >= m.stopAt {
		// land exactly on the result pocket
		for i, n := range wheel {
			if n == m.result {
				m.wheelPos = i
				break
			}
		}
		return m.settle()
	}
	// spin fast, then ease the ball into its pocket over the final frames
	interval := 1
	switch remaining := m.stopAt - m.frame; {
	case remaining <= 2:
		interval = 5
	case remaining <= 5:
		interval = 3
	case remaining <= 10:
		interval = 2
	}
	if m.frame%interval == 0 {
		m.wheelPos = (m.wheelPos + 1) % len(wheel)
	}
	return tick(45 * time.Millisecond)
}

func (m *Model) settle() tea.Cmd {
	m.phase = phaseResult
	var net int64
	for i, w := range m.wagers {
		if w == 0 {
			continue
		}
		if targets[i].wins(m.result, m.number) {
			net += w * targets[i].payout
		} else {
			net -= w
		}
	}
	m.delta = net
	switch {
	case net > 0:
		m.outcome = "Winner!"
	case net < 0:
		m.outcome = "House takes it"
	default:
		m.outcome = "Break even"
	}
	m.history = append([]int{m.result}, m.history...)
	if len(m.history) > 10 {
		m.history = m.history[:10]
	}
	return games.Settle(net)
}
