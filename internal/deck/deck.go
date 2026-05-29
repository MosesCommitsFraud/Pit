// Package deck provides a standard 52-card deck shared by the card games.
package deck

import (
	"fmt"
	"math/rand/v2"
)

// Suit is one of the four card suits.
type Suit uint8

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

// Symbol returns the unicode pip for the suit.
func (s Suit) Symbol() string {
	switch s {
	case Spades:
		return "♠"
	case Hearts:
		return "♥"
	case Diamonds:
		return "♦"
	default:
		return "♣"
	}
}

// Red reports whether the suit is rendered in red.
func (s Suit) Red() bool { return s == Hearts || s == Diamonds }

// Rank is the face value, 2..14 where 11=J, 12=Q, 13=K, 14=A.
type Rank uint8

const (
	Two Rank = iota + 2
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

// Label returns the short rank label (A, K, Q, J, 10, 9 ...).
func (r Rank) Label() string {
	switch r {
	case Ace:
		return "A"
	case King:
		return "K"
	case Queen:
		return "Q"
	case Jack:
		return "J"
	case Ten:
		return "10"
	default:
		return fmt.Sprintf("%d", uint8(r))
	}
}

// Card is a single playing card.
type Card struct {
	Rank Rank
	Suit Suit
}

// String renders the card compactly, e.g. "A♠".
func (c Card) String() string { return c.Rank.Label() + c.Suit.Symbol() }

// PokerCode returns the two-character code used by the poker eval library,
// e.g. "Ah", "Td", "2s".
func (c Card) PokerCode() string {
	rank := c.Rank.Label()
	if c.Rank == Ten {
		rank = "T"
	}
	suit := []string{"s", "h", "d", "c"}[c.Suit]
	return rank + suit
}

// Deck is a mutable, shuffleable stack of cards dealt from the top.
type Deck struct {
	cards []Card
	pos   int
}

// New returns a freshly shuffled single deck.
func New() *Deck {
	d := &Deck{}
	d.reset()
	d.Shuffle()
	return d
}

// NewShoe returns n decks shuffled together (used by blackjack).
func NewShoe(n int) *Deck {
	d := &Deck{}
	for i := 0; i < n; i++ {
		d.appendFullDeck()
	}
	d.Shuffle()
	return d
}

func (d *Deck) reset() {
	d.cards = d.cards[:0]
	d.appendFullDeck()
	d.pos = 0
}

func (d *Deck) appendFullDeck() {
	for s := Spades; s <= Clubs; s++ {
		for r := Two; r <= Ace; r++ {
			d.cards = append(d.cards, Card{Rank: r, Suit: s})
		}
	}
}

// Shuffle randomizes the remaining-and-used cards and resets the deal position.
func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
	d.pos = 0
}

// Remaining reports how many cards are left to deal.
func (d *Deck) Remaining() int { return len(d.cards) - d.pos }

// Deal returns the next card, reshuffling if the deck is exhausted.
func (d *Deck) Deal() Card {
	if d.pos >= len(d.cards) {
		d.Shuffle()
	}
	c := d.cards[d.pos]
	d.pos++
	return c
}
