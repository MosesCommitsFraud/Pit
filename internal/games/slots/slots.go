// Package slots implements a 3-reel slot machine with animated, decelerating
// reels.
package slots

import (
	"math/rand/v2"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pit/internal/econ"
	"pit/internal/games"
	"pit/internal/ui"
)

// symbol is one reel face.
type symbol struct {
	glyph  string
	color  lipgloss.Color
	weight int   // relative frequency on the strip
	pay3   int64 // multiplier when three match
}

var symbols = []symbol{
	{"C", lipgloss.Color("#d6453d"), 9, 3},  // cherry (also pays on two)
	{"♣", lipgloss.Color("#5fd07a"), 7, 4},  // clover
	{"$", lipgloss.Color("#e8c14a"), 6, 6},  // cash
	{"★", lipgloss.Color("#f3ead3"), 4, 12}, // star
	{"♦", lipgloss.Color("#62b6ff"), 3, 25}, // diamond
	{"7", lipgloss.Color("#ff5e8a"), 1, 75}, // jackpot
}

const cherryIdx = 0

// weighted strip used both for the visual cycle and for picking results.
var strip = buildStrip()

func buildStrip() []int {
	var s []int
	for i, sym := range symbols {
		for w := 0; w < sym.weight; w++ {
			s = append(s, i)
		}
	}
	return s
}

func pick() int { return strip[rand.IntN(len(strip))] }

var bets = []int64{10, 25, 50, 100, 250}

type reel struct {
	display  int // currently shown symbol index
	result   int // final symbol once stopped
	spinning bool
	frame    int
	stopAt   int
}

type phase int

const (
	phaseIdle phase = iota
	phaseSpin
	phaseResult
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(55*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Model is the slots game.
type Model struct {
	bank   *econ.Bankroll
	reels  [3]reel
	betIdx int
	phase  phase

	lastDelta int64
	lastWin   string
	width     int
	height    int
}

// New builds a slots model bound to the bankroll.
func New(b *econ.Bankroll, width, height int) tea.Model {
	m := &Model{bank: b, width: width, height: height}
	for i := range m.reels {
		m.reels[i].display = pick()
		m.reels[i].result = m.reels[i].display
	}
	// default bet: first affordable option
	for i, v := range bets {
		if b.CanCover(v) {
			m.betIdx = i
			break
		}
	}
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) bet() int64 { return bets[m.betIdx] }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "m", "esc", "q":
			return m, games.QuitToMenu
		case "left", "h":
			if m.phase != phaseSpin && m.betIdx > 0 {
				m.betIdx--
			}
		case "right", "l":
			if m.phase != phaseSpin && m.betIdx < len(bets)-1 {
				m.betIdx++
			}
		case " ", "enter":
			if m.phase != phaseSpin {
				return m, m.spin()
			}
		}

	case tickMsg:
		if m.phase == phaseSpin {
			return m, m.advance()
		}
	}
	return m, nil
}

// spin locks in random results and starts the animation.
func (m *Model) spin() tea.Cmd {
	if !m.bank.CanCover(m.bet()) {
		m.lastWin = "Not enough chips for that bet"
		return nil
	}
	m.phase = phaseSpin
	m.lastWin = ""
	m.lastDelta = 0
	for i := range m.reels {
		m.reels[i].result = pick()
		m.reels[i].spinning = true
		m.reels[i].frame = 0
		m.reels[i].stopAt = 14 + i*7 // staggered stops
	}
	return tick()
}

// advance steps the animation one frame; when all reels stop it settles.
func (m *Model) advance() tea.Cmd {
	allStopped := true
	for i := range m.reels {
		r := &m.reels[i]
		if !r.spinning {
			continue
		}
		r.frame++
		if r.frame >= r.stopAt {
			r.spinning = false
			r.display = r.result
			continue
		}
		allStopped = false
		// deceleration: advance the visible symbol less often over time
		interval := 1 + r.frame/6
		if r.frame%interval == 0 {
			r.display = (r.display + 1) % len(symbols)
		}
	}
	if allStopped {
		return m.settle()
	}
	return tick()
}

// settle evaluates the payline and emits the bankroll delta.
func (m *Model) settle() tea.Cmd {
	m.phase = phaseResult
	bet := m.bet()
	a, b, c := m.reels[0].result, m.reels[1].result, m.reels[2].result

	var ret int64
	switch {
	case a == b && b == c:
		ret = symbols[a].pay3 * bet
		m.lastWin = "THREE " + symbols[a].glyph + "!"
	case countCherries(a, b, c) >= 2:
		ret = 2 * bet
		m.lastWin = "Two cherries"
	default:
		ret = 0
	}
	m.lastDelta = ret - bet
	return games.Settle(m.lastDelta)
}

func countCherries(idx ...int) int {
	n := 0
	for _, i := range idx {
		if i == cherryIdx {
			n++
		}
	}
	return n
}

func (m *Model) View() string {
	header := ui.Header("Slots", m.bank, m.width)

	reelCells := make([]string, 3)
	for i := range m.reels {
		reelCells[i] = renderCell(m.reels[i].display)
	}
	window := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ui.Gold).
		Padding(1, 2).
		Render(lipgloss.JoinHorizontal(lipgloss.Center, reelCells...))

	payline := ui.Subtle.Render("──── payline ────")

	var status string
	switch m.phase {
	case phaseResult:
		if m.lastDelta > 0 {
			status = ui.Banner(ui.ResultWin, m.lastWin, m.lastDelta)
		} else {
			status = ui.Banner(ui.ResultLose, "No win", m.lastDelta)
		}
	default:
		if m.lastWin != "" {
			status = ui.LoseText.Render(m.lastWin)
		} else {
			status = ui.Subtle.Render("Press SPACE to spin")
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Center,
		window,
		payline,
		"",
		ui.BetSelector(m.bet()),
		"",
		status,
	)

	help := ui.Help("space spin · ←/→ bet · m menu")
	paytable := renderPaytable()

	center := lipgloss.Place(max(m.width, 1), max(m.height-4, 1),
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Center, ui.Panel.Render(body), "  ", paytable))

	return lipgloss.JoinVertical(lipgloss.Left, header, center, help)
}

func renderCell(idx int) string {
	s := symbols[idx]
	glyph := lipgloss.NewStyle().Bold(true).Foreground(s.color).Render(" " + s.glyph + " ")
	return lipgloss.NewStyle().
		Background(ui.Ink).
		Padding(1, 2).
		Margin(0, 1).
		Render(glyph)
}

func renderPaytable() string {
	var b strings.Builder
	b.WriteString(ui.Heading.Render("Pays (×bet)") + "\n")
	for i := len(symbols) - 1; i >= 0; i-- {
		s := symbols[i]
		row := lipgloss.NewStyle().Foreground(s.color).Render(s.glyph+s.glyph+s.glyph) +
			ui.Subtle.Render("  ×"+itoa(s.pay3))
		b.WriteString(row + "\n")
	}
	b.WriteString(ui.Subtle.Render("CC  ×2"))
	return ui.Panel.Render(b.String())
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var bs []byte
	for n > 0 {
		bs = append([]byte{byte('0' + n%10)}, bs...)
		n /= 10
	}
	return string(bs)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
