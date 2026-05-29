package app

import (
	"pit/internal/games"
	"pit/internal/games/blackjack"
	"pit/internal/games/booksun"
	"pit/internal/games/holdem"
	"pit/internal/games/roulette"
	"pit/internal/games/slots"
)

// Registry returns every game available in the menu, in display order.
func Registry() []games.Entry {
	return []games.Entry{
		{ID: "slots", Title: "Slots", Blurb: "Spin three reels. Chase the 7s.", New: slots.New},
		{ID: "booksun", Title: "Book of Ra", Blurb: "5 reels · expanding-symbol free spins.", New: booksun.New},
		{ID: "blackjack", Title: "Blackjack", Blurb: "Beat the dealer to 21.", New: blackjack.New},
		{ID: "roulette", Title: "Roulette", Blurb: "Bet the board, spin the wheel.", New: roulette.New},
		{ID: "holdem", Title: "Texas Hold'em", Blurb: "Outplay three bots, heads-up to showdown.", New: holdem.New},
	}
}
