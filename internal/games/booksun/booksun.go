// Package booksun implements "Book of Ra", a 5x3 Egyptian-themed slot: the Book
// is both wild and scatter, and three Books trigger ten free spins with one
// randomly chosen expanding symbol.
package booksun

import (
	"fmt"
	"math/rand/v2"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"pit/internal/econ"
	"pit/internal/games"
)

const (
	reels    = 5
	rowsN    = 3
	numLines = 10
	freeSpan = 10 // free spins awarded per trigger
)

// symbol indices
const (
	symTen = iota
	symJack
	symQueen
	symKing
	symAce
	symScarab
	symIdol
	symMask
	symHero
	book // wild + scatter
)

type symInfo struct {
	glyph  string // single Egyptian-themed icon shown on the reels
	name   string // label for the paytable
	weight int
	pay    [6]int64 // pay[k] = line-bet multiplier for k-of-a-kind (k 0..5)
}

var syms = []symInfo{
	symTen:    {"10", "10", 14, [6]int64{0, 0, 0, 5, 20, 100}},
	symJack:   {"J", "J", 13, [6]int64{0, 0, 0, 5, 20, 100}},
	symQueen:  {"Q", "Q", 12, [6]int64{0, 0, 0, 5, 25, 100}},
	symKing:   {"K", "K", 11, [6]int64{0, 0, 0, 5, 30, 150}},
	symAce:    {"A", "A", 10, [6]int64{0, 0, 0, 5, 30, 150}},
	symScarab: {"✦", "SCARAB", 8, [6]int64{0, 0, 0, 5, 40, 400}},
	symIdol:   {"▲", "PYRAMID", 7, [6]int64{0, 0, 0, 5, 40, 400}},
	symMask:   {"♛", "PHARAOH", 5, [6]int64{0, 0, 0, 10, 100, 500}},
	symHero:   {"☥", "EXPLORER", 4, [6]int64{0, 0, 2, 20, 200, 1000}},
	book:      {"☉", "BOOK", 3, [6]int64{0, 0, 0, 0, 0, 0}},
}

// bookScatter pays a multiple of the TOTAL bet for k books anywhere.
var bookScatter = map[int]int64{3: 2, 4: 20, 5: 200}

// paylines as a row index (0..2) per reel.
var paylines = [numLines][reels]int{
	{1, 1, 1, 1, 1},
	{0, 0, 0, 0, 0},
	{2, 2, 2, 2, 2},
	{0, 1, 2, 1, 0},
	{2, 1, 0, 1, 2},
	{0, 0, 1, 2, 2},
	{2, 2, 1, 0, 0},
	{1, 0, 0, 0, 1},
	{1, 2, 2, 2, 1},
	{0, 1, 1, 1, 0},
}

// total stakes; per-line bet is stake / numLines.
var bets = []int64{10, 20, 50, 100, 250}

var strip = buildStrip()

func buildStrip() []int {
	var s []int
	for i, sy := range syms {
		for w := 0; w < sy.weight; w++ {
			s = append(s, i)
		}
	}
	return s
}

func pick() int { return strip[rand.IntN(len(strip))] }

// symPay returns the line-bet multiplier for k of symbol s.
func symPay(s, k int) int64 {
	if k < 0 || k > 5 {
		return 0
	}
	return syms[s].pay[k]
}

type phase int

const (
	phaseReady phase = iota
	phaseSpin
	phaseResult
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(45*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Model is the Book of Ra slot.
type Model struct {
	bank   *econ.Bankroll
	betIdx int

	grid   [reels][rowsN]int // currently displayed
	result [reels][rowsN]int // locked-in spin outcome

	frame    [reels]int
	stopAt   [reels]int
	spinning [reels]bool
	phase    phase

	inFree   bool
	freeLeft int
	special  int

	winLines []int
	winCells map[[2]int]bool
	expanded []int // reels the special expanded across (free spins)
	lastWin  int64
	msg      string

	width, height int
}

// New builds a Book of Ra model bound to the bankroll.
func New(b *econ.Bankroll, width, height int) tea.Model {
	m := &Model{bank: b, width: width, height: height, winCells: map[[2]int]bool{}}
	for r := 0; r < reels; r++ {
		for row := 0; row < rowsN; row++ {
			m.grid[r][row] = pick()
		}
	}
	for i, v := range bets {
		if b.CanCover(v) {
			m.betIdx = i
			break
		}
	}
	m.msg = "press SPACE to spin"
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) stake() int64   { return bets[m.betIdx] }
func (m *Model) lineBet() int64 { return bets[m.betIdx] / numLines }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "m", "esc", "q":
			return m, games.QuitToMenu
		case "left", "h":
			if m.phase != phaseSpin && !m.inFree && m.betIdx > 0 {
				m.betIdx--
			}
		case "right", "l":
			if m.phase != phaseSpin && !m.inFree && m.betIdx < len(bets)-1 {
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

func (m *Model) spin() tea.Cmd {
	if !m.inFree && !m.bank.CanCover(m.stake()) {
		m.msg = "not enough chips for that bet"
		return nil
	}
	if m.inFree {
		m.freeLeft--
	}
	m.phase = phaseSpin
	m.lastWin = 0
	m.winLines = nil
	m.winCells = map[[2]int]bool{}
	m.expanded = nil
	m.msg = ""
	for r := 0; r < reels; r++ {
		for row := 0; row < rowsN; row++ {
			m.result[r][row] = pick()
		}
		m.spinning[r] = true
		m.frame[r] = 0
		m.stopAt[r] = 26 + r*9 // staggered, left-to-right
	}
	return tick()
}

func (m *Model) advance() tea.Cmd {
	allStopped := true
	for r := 0; r < reels; r++ {
		if !m.spinning[r] {
			continue
		}
		m.frame[r]++
		if m.frame[r] >= m.stopAt[r] {
			m.spinning[r] = false
			for row := 0; row < rowsN; row++ {
				m.grid[r][row] = m.result[r][row]
			}
			continue
		}
		allStopped = false
		interval := 1
		switch remaining := m.stopAt[r] - m.frame[r]; {
		case remaining <= 2:
			interval = 4
		case remaining <= 4:
			interval = 3
		case remaining <= 7:
			interval = 2
		}
		if m.frame[r]%interval == 0 {
			for row := 0; row < rowsN; row++ {
				m.grid[r][row] = pick()
			}
		}
	}
	if allStopped {
		return m.settle()
	}
	return tick()
}

func (m *Model) settle() tea.Cmd {
	m.phase = phaseResult

	lineWin, lines, cells := evalLines(m.result, m.lineBet())
	m.winLines, m.winCells = lines, cells

	books := countBooks(m.result)
	var scatterWin int64
	if p, ok := bookScatter[clampBooks(books)]; ok {
		scatterWin = p * m.stake()
	}

	win := lineWin + scatterWin
	wasFree := m.inFree

	// expanding symbol pays only during free spins
	if m.inFree {
		expWin, expReels := evalExpand(m.result, m.special, m.lineBet())
		win += expWin
		m.expanded = expReels
		for _, r := range expReels {
			for row := 0; row < rowsN; row++ {
				m.winCells[[2]int{r, row}] = true
			}
		}
	}

	var net int64
	if wasFree {
		net = win // free spins cost nothing
	} else {
		net = win - m.stake()
	}
	m.lastWin = win

	// free-spin trigger / retrigger
	triggered := false
	if books >= 3 {
		if m.inFree {
			m.freeLeft += freeSpan
		} else {
			m.inFree = true
			m.freeLeft = freeSpan
			m.special = pickSpecial()
			triggered = true
		}
	}

	m.msg = m.statusMessage(win, wasFree, triggered, books)

	// end of the free-spins round
	if m.inFree && m.freeLeft <= 0 && !triggered {
		m.inFree = false
		m.msg = "free spins over"
	}

	return games.Settle(net)
}

func (m *Model) statusMessage(win int64, wasFree, triggered bool, books int) string {
	switch {
	case triggered:
		return fmt.Sprintf("%d BOOKS — %d FREE SPINS! expanding symbol: %s %s", books, freeSpan, syms[m.special].glyph, syms[m.special].name)
	case wasFree && books >= 3:
		return fmt.Sprintf("%d BOOKS — +%d FREE SPINS!", books, freeSpan)
	case win > 0:
		if wasFree {
			return "FREE SPIN WIN"
		}
		return "WIN"
	default:
		if wasFree {
			return "no win"
		}
		return "no win"
	}
}

func clampBooks(n int) int {
	if n > 5 {
		return 5
	}
	return n
}

func countBooks(g [reels][rowsN]int) int {
	n := 0
	for r := 0; r < reels; r++ {
		for row := 0; row < rowsN; row++ {
			if g[r][row] == book {
				n++
			}
		}
	}
	return n
}

// evalLines scores every payline left-to-right (Book is wild) and returns the
// total win, the winning line indices, and the set of winning cells.
func evalLines(g [reels][rowsN]int, lineBet int64) (int64, []int, map[[2]int]bool) {
	var total int64
	var wins []int
	cells := map[[2]int]bool{}

	for li := 0; li < numLines; li++ {
		line := paylines[li]
		var seq [reels]int
		for r := 0; r < reels; r++ {
			seq[r] = g[r][line[r]]
		}
		// target symbol: first non-book (book is wild)
		target := seq[0]
		if target == book {
			for r := 1; r < reels; r++ {
				if seq[r] != book {
					target = seq[r]
					break
				}
			}
		}
		// run of target-or-wild from the left
		run := 0
		for r := 0; r < reels; r++ {
			if seq[r] == target || seq[r] == book {
				run++
			} else {
				break
			}
		}
		if pay := symPay(target, run) * lineBet; pay > 0 {
			total += pay
			wins = append(wins, li)
			for r := 0; r < run; r++ {
				cells[[2]int{r, line[r]}] = true
			}
		}
	}
	return total, wins, cells
}

// evalExpand handles the free-spin expanding symbol: if the special appears on
// enough reels it expands across them and pays on all lines.
func evalExpand(g [reels][rowsN]int, special int, lineBet int64) (int64, []int) {
	var reelsWith []int
	for r := 0; r < reels; r++ {
		for row := 0; row < rowsN; row++ {
			if g[r][row] == special {
				reelsWith = append(reelsWith, r)
				break
			}
		}
	}
	min := 3
	if special == symHero {
		min = 2
	}
	if len(reelsWith) < min {
		return 0, nil
	}
	// expanded symbol pays its k-of-a-kind value on every line
	return symPay(special, len(reelsWith)) * lineBet * numLines, reelsWith
}

// pickSpecial chooses the expanding symbol for a free-spins round.
func pickSpecial() int { return rand.IntN(symHero + 1) } // 0..symHero (any paying symbol)
