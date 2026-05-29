package booksun

import "testing"

// grid helper: rows are given top(0), middle(1), bottom(2), each length 5.
func makeGrid(top, mid, bot [reels]int) [reels][rowsN]int {
	var g [reels][rowsN]int
	for r := 0; r < reels; r++ {
		g[r][0] = top[r]
		g[r][1] = mid[r]
		g[r][2] = bot[r]
	}
	return g
}

// non-paying filler rows for top/bottom so only the middle line can win.
var fillTop = [reels]int{symTen, symJack, symTen, symJack, symTen}
var fillBot = [reels]int{symQueen, symKing, symQueen, symKing, symQueen}

func TestLinePaysThreeOfAKind(t *testing.T) {
	mid := [reels]int{symHero, symHero, symHero, symTen, symJack}
	g := makeGrid(fillTop, mid, fillBot)

	total, lines, _ := evalLines(g, 1)
	want := symPay(symHero, 3) // 20
	if total != want {
		t.Fatalf("total = %d want %d", total, want)
	}
	if len(lines) != 1 || lines[0] != 0 {
		t.Fatalf("winning lines = %v want [0]", lines)
	}
}

func TestBookActsAsWildInLine(t *testing.T) {
	mid := [reels]int{symHero, book, symHero, symTen, symJack}
	g := makeGrid(fillTop, mid, fillBot)

	total, _, _ := evalLines(g, 1)
	if want := symPay(symHero, 3); total != want {
		t.Fatalf("wild substitution: total = %d want %d", total, want)
	}
}

func TestScatterCountTriggers(t *testing.T) {
	mid := [reels]int{book, symTen, book, symJack, book}
	g := makeGrid(fillTop, mid, fillBot)
	if n := countBooks(g); n != 3 {
		t.Fatalf("book count = %d want 3", n)
	}
	if _, ok := bookScatter[3]; !ok {
		t.Fatal("3 books should pay a scatter")
	}
}

func TestExpandPaysAcrossLines(t *testing.T) {
	// HERO on reels 0,1,2 (>= 2 reels) expands and pays on all lines.
	mid := [reels]int{symHero, symHero, symHero, symTen, symJack}
	g := makeGrid(fillTop, mid, fillBot)

	win, reelsWith := evalExpand(g, symHero, 1)
	if len(reelsWith) != 3 {
		t.Fatalf("expanded reels = %v want 3", reelsWith)
	}
	if want := symPay(symHero, 3) * numLines; win != want {
		t.Fatalf("expand win = %d want %d", win, want)
	}
}

func TestExpandNeedsEnoughReels(t *testing.T) {
	// A non-hero symbol needs >= 3 reels; only 2 here means no expand.
	mid := [reels]int{symMask, symMask, symTen, symJack, symTen}
	g := makeGrid(fillTop, mid, fillBot)
	if win, _ := evalExpand(g, symMask, 1); win != 0 {
		t.Fatalf("expand with 2 reels (non-hero) should not pay, got %d", win)
	}
}
