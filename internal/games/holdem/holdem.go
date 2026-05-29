// Package holdem implements single-table Texas hold'em against bot opponents,
// with blinds, four betting streets, and a poker-library showdown.
package holdem

import (
	"fmt"
	"math/rand/v2"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"pit/internal/deck"
	"pit/internal/econ"
	"pit/internal/games"
)

const (
	smallBlind  int64 = 10
	bigBlind    int64 = 20
	botStack    int64 = 1000
	botDelay          = 520 * time.Millisecond
	runoutDelay       = 620 * time.Millisecond
)

// betting streets
const (
	preflop = iota
	flop
	turn
	river
)

type seat struct {
	name    string
	hole    []deck.Card
	stack   int64
	street  int64 // chips committed on the current street
	total   int64 // chips committed this hand
	folded  bool
	allIn   bool
	isHuman bool
	lastAct string // for display: "calls", "raises", etc.
}

type phase int

const (
	phaseIdle phase = iota // between hands / not enough chips
	phaseHuman
	phaseBot
	phaseRunout
	phaseShowdown
)

type tickMsg time.Time

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Model is the hold'em game.
type Model struct {
	bank  *econ.Bankroll
	d     *deck.Deck
	seats []seat
	board []deck.Card

	pot        int64
	button     int
	toAct      int
	street     int
	currentBet int64
	lastRaise  int64
	acted      []bool

	phase     phase
	startBank int64 // human stack at the hand's start
	delta     int64 // settled net for the finished hand
	log       string
	result    string

	raiseSizes []int64
	raiseIdx   int

	width, height int
}

// New builds a hold'em model bound to the bankroll with three bot opponents.
func New(b *econ.Bankroll, width, height int) tea.Model {
	m := &Model{
		bank:   b,
		d:      deck.New(),
		width:  width,
		height: height,
		seats: []seat{
			{name: "You", isHuman: true},
			{name: "Ada"},
			{name: "Boyd"},
			{name: "Cleo"},
		},
		button: 0,
	}
	m.acted = make([]bool, len(m.seats))
	m.phase = phaseIdle
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) human() *seat { return &m.seats[0] }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tickMsg:
		switch m.phase {
		case phaseBot:
			return m, m.runBot()
		case phaseRunout:
			return m, m.dealStreet()
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "m", "esc", "q":
			return m, games.QuitToMenu
		}
		switch m.phase {
		case phaseIdle, phaseShowdown:
			if msg.String() == "enter" || msg.String() == " " {
				return m, m.startHand()
			}
		case phaseHuman:
			return m.humanKey(msg)
		}
	}
	return m, nil
}

// startHand resets the table, posts blinds and deals hole cards.
func (m *Model) startHand() tea.Cmd {
	if m.bank.Balance() < bigBlind {
		m.phase = phaseIdle
		m.result = "Not enough chips to post the big blind"
		return nil
	}
	m.d.Shuffle()
	m.board = nil
	m.pot = 0
	m.street = preflop
	m.currentBet = 0
	m.lastRaise = bigBlind
	m.result = ""
	m.delta = 0
	m.startBank = m.bank.Balance()
	for i := range m.seats {
		s := &m.seats[i]
		s.hole = nil
		s.street, s.total = 0, 0
		s.folded, s.allIn = false, false
		s.lastAct = ""
		if s.isHuman {
			s.stack = m.bank.Balance()
		} else {
			s.stack = botStack
		}
		m.acted[i] = false
	}
	// deal two hole cards each
	for r := 0; r < 2; r++ {
		for i := range m.seats {
			m.seats[i].hole = append(m.seats[i].hole, m.d.Deal())
		}
	}
	// post blinds: SB left of button, BB next
	sb := m.next(m.button)
	bb := m.next(sb)
	m.postBlind(sb, smallBlind)
	m.postBlind(bb, bigBlind)
	m.currentBet = bigBlind
	// preflop action starts left of the big blind
	m.toAct = m.next(bb)
	m.log = "Blinds posted"
	return m.setActorPhase()
}

func (m *Model) postBlind(i int, amt int64) {
	paid := m.commit(i, amt)
	m.seats[i].lastAct = "blind"
	_ = paid
}

// commit moves up to amt chips from seat i into its street commitment.
func (m *Model) commit(i int, amt int64) int64 {
	s := &m.seats[i]
	if amt > s.stack {
		amt = s.stack
	}
	s.stack -= amt
	s.street += amt
	s.total += amt
	if s.stack == 0 {
		s.allIn = true
	}
	return amt
}

func (m *Model) next(i int) int { return (i + 1) % len(m.seats) }

func (m *Model) contenders() int {
	n := 0
	for i := range m.seats {
		if !m.seats[i].folded {
			n++
		}
	}
	return n
}

func (m *Model) needsAction(i int) bool {
	s := &m.seats[i]
	return !s.folded && !s.allIn && (!m.acted[i] || s.street < m.currentBet)
}

// nextToAct returns the next seat that must act after `from`, or -1.
func (m *Model) nextToAct(from int) int {
	for k := 1; k <= len(m.seats); k++ {
		i := (from + k) % len(m.seats)
		if m.needsAction(i) {
			return i
		}
	}
	return -1
}

// setActorPhase configures the phase for whoever is to act and returns the
// command to drive it.
func (m *Model) setActorPhase() tea.Cmd {
	if m.seats[m.toAct].isHuman {
		m.phase = phaseHuman
		m.computeRaiseSizes()
		return nil
	}
	m.phase = phaseBot
	return tick(botDelay)
}

// ---- action application -------------------------------------------------

func (m *Model) callAmount(i int) int64 {
	c := m.currentBet - m.seats[i].street
	if c < 0 {
		return 0
	}
	return c
}

func (m *Model) doFold(i int) {
	m.seats[i].folded = true
	m.seats[i].lastAct = "folds"
	m.acted[i] = true
}

func (m *Model) doCallCheck(i int) {
	c := m.callAmount(i)
	if c == 0 {
		m.seats[i].lastAct = "checks"
	} else {
		paid := m.commit(i, c)
		if m.seats[i].allIn {
			m.seats[i].lastAct = "all-in"
		} else {
			m.seats[i].lastAct = fmt.Sprintf("calls %d", paid)
		}
	}
	m.acted[i] = true
}

// doRaiseTo raises seat i's street commitment to target (clamped to all-in).
func (m *Model) doRaiseTo(i int, target int64) {
	s := &m.seats[i]
	maxTarget := s.street + s.stack
	if target > maxTarget {
		target = maxTarget
	}
	add := target - s.street
	m.commit(i, add)
	raiseBy := s.street - m.currentBet
	if raiseBy > 0 {
		m.lastRaise = raiseBy
	}
	m.currentBet = s.street
	if s.allIn {
		s.lastAct = "all-in"
	} else {
		s.lastAct = fmt.Sprintf("raises to %d", s.street)
	}
	// reopen action for everyone still in
	for j := range m.seats {
		if j != i && !m.seats[j].folded && !m.seats[j].allIn {
			m.acted[j] = false
		}
	}
	m.acted[i] = true
}

// afterAction advances the hand after the seat at m.toAct has acted.
func (m *Model) afterAction() tea.Cmd {
	if m.contenders() == 1 {
		return m.awardFold()
	}
	nxt := m.nextToAct(m.toAct)
	if nxt == -1 {
		return m.closeRound()
	}
	m.toAct = nxt
	return m.setActorPhase()
}

func (m *Model) closeRound() tea.Cmd {
	for i := range m.seats {
		m.pot += m.seats[i].street
		m.seats[i].street = 0
	}
	if m.street == river {
		return m.showdown()
	}
	return m.dealStreet()
}

// dealStreet advances to and deals the next community street, then either lets
// the next player act, runs out remaining streets, or goes to showdown.
func (m *Model) dealStreet() tea.Cmd {
	m.street++
	switch m.street {
	case flop:
		m.board = append(m.board, m.d.Deal(), m.d.Deal(), m.d.Deal())
	case turn, river:
		m.board = append(m.board, m.d.Deal())
	}
	m.currentBet = 0
	m.lastRaise = bigBlind
	for i := range m.seats {
		m.seats[i].street = 0
		if !m.seats[i].folded {
			m.seats[i].lastAct = ""
		}
		m.acted[i] = false
	}
	nxt := m.nextToAct(m.button)
	if nxt == -1 {
		// no one can act (all-in): run the board out with a delay
		if m.street == river {
			return m.showdown()
		}
		m.phase = phaseRunout
		return tick(runoutDelay)
	}
	m.toAct = nxt
	return m.setActorPhase()
}

// ---- endings ------------------------------------------------------------

func (m *Model) awardFold() tea.Cmd {
	for i := range m.seats {
		m.pot += m.seats[i].street
		m.seats[i].street = 0
	}
	winner := -1
	for i := range m.seats {
		if !m.seats[i].folded {
			winner = i
			break
		}
	}
	m.seats[winner].stack += m.pot
	if m.seats[winner].isHuman {
		m.result = fmt.Sprintf("Everyone folded — you win %d", m.pot)
	} else {
		m.result = fmt.Sprintf("%s wins %d (all folded)", m.seats[winner].name, m.pot)
	}
	return m.finish()
}

func (m *Model) showdown() tea.Cmd {
	best := int32(maxRank + 1)
	var winners []int
	var bestName string
	for i := range m.seats {
		if m.seats[i].folded {
			continue
		}
		r := rankOf(m.seats[i].hole, m.board)
		if r < best {
			best = r
			winners = winners[:0]
			winners = append(winners, i)
			bestName = rankName(r)
		} else if r == best {
			winners = append(winners, i)
		}
	}
	share := m.pot / int64(len(winners))
	rem := m.pot - share*int64(len(winners))
	for k, i := range winners {
		s := share
		if int64(k) < rem { // give odd chips to first winners
			s++
		}
		m.seats[i].stack += s
	}
	names := ""
	for k, i := range winners {
		if k > 0 {
			names += ", "
		}
		names += m.seats[i].name
	}
	m.result = fmt.Sprintf("%s win%s with %s", names, plural(len(winners)), bestName)
	m.button = m.next(m.button)
	return m.finish()
}

func plural(n int) string {
	if n == 1 {
		return "s"
	}
	return ""
}

// finish settles the human's net change for the hand and emits it.
func (m *Model) finish() tea.Cmd {
	m.phase = phaseShowdown
	m.delta = m.human().stack - m.startBank
	// reveal everyone for the showdown
	return games.Settle(m.delta)
}

// ---- human input --------------------------------------------------------

func (m *Model) humanKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	i := m.toAct
	switch msg.String() {
	case "f":
		m.doFold(i)
		return m, m.afterAction()
	case "c": // check or call
		m.doCallCheck(i)
		return m, m.afterAction()
	case "left", "h":
		if m.raiseIdx > 0 {
			m.raiseIdx--
		}
	case "right", "l":
		if m.raiseIdx < len(m.raiseSizes)-1 {
			m.raiseIdx++
		}
	case "r", "enter": // raise/bet selected size
		if len(m.raiseSizes) > 0 {
			m.doRaiseTo(i, m.raiseSizes[m.raiseIdx])
			return m, m.afterAction()
		}
	}
	return m, nil
}

// computeRaiseSizes builds the legal raise targets offered to the human.
func (m *Model) computeRaiseSizes() {
	s := m.human()
	maxTarget := s.street + s.stack // all-in
	minTarget := m.currentBet + m.minRaiseStep()
	m.raiseSizes = m.raiseSizes[:0]
	if maxTarget <= m.currentBet {
		m.raiseIdx = 0
		return // can't raise (no chips)
	}
	if minTarget > maxTarget {
		minTarget = maxTarget // only all-in available
	}
	add := func(t int64) {
		if t < minTarget {
			t = minTarget
		}
		if t > maxTarget {
			t = maxTarget
		}
		for _, e := range m.raiseSizes {
			if e == t {
				return
			}
		}
		m.raiseSizes = append(m.raiseSizes, t)
	}
	pot := m.pot + m.tableStreet()
	add(minTarget)
	add(m.currentBet + pot/2)
	add(m.currentBet + pot)
	add(maxTarget)
	m.raiseIdx = 0
}

func (m *Model) minRaiseStep() int64 {
	if m.lastRaise > bigBlind {
		return m.lastRaise
	}
	return bigBlind
}

func (m *Model) tableStreet() int64 {
	var t int64
	for i := range m.seats {
		t += m.seats[i].street
	}
	return t
}

// ---- bot ---------------------------------------------------------------

func (m *Model) runBot() tea.Cmd {
	i := m.toAct
	s := &m.seats[i]
	toCall := m.callAmount(i)

	var strength float64
	if len(m.board) == 0 {
		strength = preflopStrength(s.hole)
	} else {
		strength = strengthFromRank(rankOf(s.hole, m.board))
	}
	strength += (rand.Float64() - 0.5) * 0.12

	canRaise := s.street+s.stack > m.currentBet
	pot := m.pot + m.tableStreet()

	switch {
	case strength > 0.82 && canRaise:
		target := m.currentBet + max64(m.minRaiseStep(), pot/2)
		m.doRaiseTo(i, target)
	case strength > 0.5:
		if toCall == 0 {
			if canRaise && rand.Float64() < 0.25 {
				m.doRaiseTo(i, m.currentBet+max64(m.minRaiseStep(), pot/3))
			} else {
				m.doCallCheck(i)
			}
		} else if float64(toCall) <= float64(s.stack)*0.4 {
			m.doCallCheck(i)
		} else {
			m.doFold(i)
		}
	default:
		if toCall == 0 {
			m.doCallCheck(i) // free check
		} else if canRaise && rand.Float64() < 0.07 {
			m.doRaiseTo(i, m.currentBet+max64(m.minRaiseStep(), pot/2)) // bluff
		} else {
			m.doFold(i)
		}
	}
	return m.afterAction()
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
