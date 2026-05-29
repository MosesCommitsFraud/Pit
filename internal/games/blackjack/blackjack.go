// Package blackjack implements blackjack vs the house: hit, stand, double, and
// split against a dealer that stands on 17, with 3:2 naturals.
package blackjack

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"pit/internal/deck"
	"pit/internal/econ"
	"pit/internal/games"
)

var bets = []int64{10, 25, 50, 100, 250}

type phase int

const (
	phaseBetting phase = iota
	phaseDealing
	phasePlayer
	phaseDealer
	phaseResult
)

type hand struct {
	cards   []deck.Card
	bet     int64
	doubled bool
	done    bool
	bust    bool
}

type tickMsg time.Time

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Model is the blackjack game.
type Model struct {
	bank   *econ.Bankroll
	shoe   *deck.Deck
	betIdx int

	dealer   []deck.Card
	holeDown bool
	players  []hand
	active   int
	wasSplit bool
	reveal   int // cards flipped during the deal animation (0..4)

	phase   phase
	outcome string
	delta   int64

	width, height int
}

// New builds a blackjack model bound to the bankroll.
func New(b *econ.Bankroll, width, height int) tea.Model {
	m := &Model{
		bank:   b,
		shoe:   deck.NewShoe(6),
		width:  width,
		height: height,
		phase:  phaseBetting,
	}
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
	case tickMsg:
		return m.onTick()
	case tea.KeyMsg:
		switch msg.String() {
		case "m", "esc", "q":
			return m, games.QuitToMenu
		}
		switch m.phase {
		case phaseBetting, phaseResult:
			return m.updateBetting(msg)
		case phasePlayer:
			return m.updatePlayer(msg)
		}
	}
	return m, nil
}

func (m *Model) updateBetting(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		if m.betIdx > 0 {
			m.betIdx--
		}
	case "right", "l":
		if m.betIdx < len(bets)-1 {
			m.betIdx++
		}
	case "enter", " ":
		if m.bank.CanCover(m.bet()) {
			return m, m.deal()
		}
		m.outcome = "Not enough chips"
	}
	return m, nil
}

func (m *Model) updatePlayer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	h := &m.players[m.active]
	switch msg.String() {
	case "h": // hit
		h.cards = append(h.cards, m.shoe.Deal())
		if v, _ := handValue(h.cards); v > 21 {
			h.bust = true
			h.done = true
			return m, m.advance()
		}
	case "s": // stand
		h.done = true
		return m, m.advance()
	case "d": // double
		if len(h.cards) == 2 && m.bank.CanCover(m.committed()+h.bet) {
			h.doubled = true
			h.bet *= 2
			h.cards = append(h.cards, m.shoe.Deal())
			if v, _ := handValue(h.cards); v > 21 {
				h.bust = true
			}
			h.done = true
			return m, m.advance()
		}
	case "p": // split
		if m.canSplit(h) {
			m.split()
		}
	}
	return m, nil
}

// committed sums the chips already at risk across all hands (for affordability).
func (m *Model) committed() int64 {
	var t int64
	for _, h := range m.players {
		t += h.bet
	}
	return t
}

func (m *Model) canSplit(h *hand) bool {
	return len(m.players) < 4 &&
		len(h.cards) == 2 &&
		h.cards[0].Rank == h.cards[1].Rank &&
		m.bank.CanCover(m.committed()+m.bet())
}

func (m *Model) split() {
	h := &m.players[m.active]
	moved := h.cards[1]
	h.cards = h.cards[:1]
	h.cards = append(h.cards, m.shoe.Deal())
	nh := hand{cards: []deck.Card{moved, m.shoe.Deal()}, bet: m.bet()}
	// insert the new hand right after the active one
	m.players = append(m.players, hand{})
	copy(m.players[m.active+2:], m.players[m.active+1:])
	m.players[m.active+1] = nh
	m.wasSplit = true
}

// deal sets up a fresh round and starts the dealing animation.
func (m *Model) deal() tea.Cmd {
	m.players = []hand{{cards: nil, bet: m.bet()}}
	m.dealer = nil
	m.active = 0
	m.wasSplit = false
	m.holeDown = true
	m.reveal = 0
	m.outcome = ""
	m.delta = 0
	// pre-deal the four cards
	m.players[0].cards = []deck.Card{m.shoe.Deal()}
	m.dealer = []deck.Card{m.shoe.Deal()}
	m.players[0].cards = append(m.players[0].cards, m.shoe.Deal())
	m.dealer = append(m.dealer, m.shoe.Deal())
	m.phase = phaseDealing
	return tick(130 * time.Millisecond)
}

func (m *Model) onTick() (tea.Model, tea.Cmd) {
	switch m.phase {
	case phaseDealing:
		m.reveal++
		if m.reveal < 4 {
			return m, tick(130 * time.Millisecond)
		}
		return m, m.afterDeal()
	case phaseDealer:
		// reveal hole, then draw to 17
		if m.holeDown {
			m.holeDown = false
			return m, tick(360 * time.Millisecond)
		}
		if v, _ := handValue(m.dealer); v < 17 {
			m.dealer = append(m.dealer, m.shoe.Deal())
			return m, tick(360 * time.Millisecond)
		}
		return m, m.resolve()
	}
	return m, nil
}

// afterDeal checks for naturals and either resolves or starts the player turn.
func (m *Model) afterDeal() tea.Cmd {
	pv, _ := handValue(m.players[0].cards)
	dv, _ := handValue(m.dealer)
	playerBJ := pv == 21
	dealerBJ := dv == 21
	if playerBJ || dealerBJ {
		m.holeDown = false
		return m.resolve()
	}
	m.phase = phasePlayer
	return nil
}

// advance moves to the next unfinished hand or to the dealer.
func (m *Model) advance() tea.Cmd {
	for i := m.active + 1; i < len(m.players); i++ {
		if !m.players[i].done {
			m.active = i
			return nil
		}
	}
	// all player hands resolved → dealer's turn (skip if everyone busted)
	allBust := true
	for _, h := range m.players {
		if !h.bust {
			allBust = false
		}
	}
	m.phase = phaseDealer
	if allBust {
		m.holeDown = false
		return m.resolve()
	}
	return tick(360 * time.Millisecond)
}

// resolve scores every hand against the dealer and emits the net delta.
func (m *Model) resolve() tea.Cmd {
	m.phase = phaseResult
	dv, _ := handValue(m.dealer)
	dealerBust := dv > 21
	dealerBJ := len(m.dealer) == 2 && dv == 21

	var net int64
	wins, losses, pushes := 0, 0, 0
	for _, h := range m.players {
		pv, _ := handValue(h.cards)
		playerBJ := !m.wasSplit && len(h.cards) == 2 && pv == 21
		switch {
		case h.bust:
			net -= h.bet
			losses++
		case playerBJ && dealerBJ:
			pushes++
		case playerBJ:
			net += h.bet * 3 / 2 // 3:2
			wins++
		case dealerBJ:
			net -= h.bet
			losses++
		case dealerBust || pv > dv:
			net += h.bet
			wins++
		case pv < dv:
			net -= h.bet
			losses++
		default:
			pushes++
		}
	}
	m.delta = net
	m.outcome = summarize(wins, losses, pushes, net)
	return games.Settle(net)
}

func summarize(w, l, p int, net int64) string {
	switch {
	case net > 0:
		return fmt.Sprintf("You win  (%dW %dL %dP)", w, l, p)
	case net < 0:
		return fmt.Sprintf("You lose  (%dW %dL %dP)", w, l, p)
	default:
		return "Push"
	}
}

func cardValue(r deck.Rank) int {
	switch {
	case r == deck.Ace:
		return 11
	case r >= deck.Ten: // Ten, Jack, Queen, King
		return 10
	default:
		return int(r)
	}
}

// handValue returns the best total and whether it is soft (an ace counts as 11).
func handValue(cards []deck.Card) (total int, soft bool) {
	aces := 0
	for _, c := range cards {
		total += cardValue(c.Rank)
		if c.Rank == deck.Ace {
			aces++
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total, aces > 0
}
