package holdem

import (
	"github.com/chehsunliu/poker"
	"pit/internal/deck"
)

// maxRank is the worst (highest) value Evaluate can return; rank 1 is the best.
const maxRank = 7462

func toPoker(cards []deck.Card) []poker.Card {
	out := make([]poker.Card, len(cards))
	for i, c := range cards {
		out[i] = poker.NewCard(c.PokerCode())
	}
	return out
}

// rankOf returns the 7-card hand rank (lower is stronger). Requires len(hole)+
// len(board) >= 5.
func rankOf(hole, board []deck.Card) int32 {
	all := make([]deck.Card, 0, len(hole)+len(board))
	all = append(all, hole...)
	all = append(all, board...)
	return poker.Evaluate(toPoker(all))
}

// rankName returns a human label like "Full House".
func rankName(r int32) string { return poker.RankString(r) }

// strength maps a made-hand rank to 0..1 (1 = nuts).
func strengthFromRank(r int32) float64 {
	return 1 - float64(r-1)/float64(maxRank-1)
}

// preflopStrength approximates two-card preflop strength on a 0..1 scale.
func preflopStrength(h []deck.Card) float64 {
	hi, lo := int(h[0].Rank), int(h[1].Rank)
	if lo > hi {
		hi, lo = lo, hi
	}
	s := float64(hi) / 14 * 0.45
	s += float64(lo) / 14 * 0.15
	if h[0].Rank == h[1].Rank { // pocket pair
		s += 0.30 + float64(hi)/14*0.1
	}
	if h[0].Suit == h[1].Suit {
		s += 0.07
	}
	if gap := hi - lo; h[0].Rank != h[1].Rank && gap <= 2 {
		s += 0.05
	}
	if s > 1 {
		s = 1
	}
	return s
}
