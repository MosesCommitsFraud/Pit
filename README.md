# ♠ Pit — a terminal casino

Pit is an animated TUI casino written in Go with the [Charm](https://charm.sh)
stack. Spin the slots, work the roulette board, beat the dealer at blackjack, or
take three bots to showdown in Texas hold'em — all without leaving your
terminal.

A retro, high-contrast look: monochrome grayscale with a single crimson accent
for wins and the red card suits.

```
▌ PIT · BLACKJACK ▐                              CAREER   $2,140
════════════════════════════════════════════════════════════════
  DEALER  20
  [  Q♣ ] [ 10♠ ]

  YOU
  [ 10♦ ] [  8♥ ]
  18

  BET  $50   ‹ ›

  YOU WIN  +$50
════════════════════════════════════════════════════════════════
 H hit · S stand · D double · P split            M menu
```

## Games

| Game            | Highlights                                                        |
| --------------- | ----------------------------------------------------------------- |
| **Slots**       | 3 weighted reels with decelerating spin animation and a paytable. |
| **Blackjack**   | 6-deck shoe, hit/stand/double/split, dealer stands on 17, 3:2 BJ. |
| **Roulette**    | European single-zero wheel, animated ball, inside & outside bets. |
| **Texas Hold'em** | Heads-up vs 3 bots: blinds, four betting streets, real showdowns. |

## Modes

- **Sandbox** — unlimited play money (starts at $10,000), resets each launch.
  For the love of the game.
- **Career** — a persistent bankroll. You collect **1,000 chips every day**, and
  anything you win on top is kept. Bust to zero and the daily stipend still
  arrives tomorrow. Saved to `~/.config/pit/save.db` (SQLite).

## Run

Requires Go 1.24+.

```sh
go run .            # from the repo root
# or build a static binary:
go build -o pit . && ./pit
```

### Controls

- **Menu:** `↑/↓` move · `enter` select · `esc` back · `ctrl+c` quit
- **Slots:** `space` spin · `←/→` bet
- **Blackjack:** `h` hit · `s` stand · `d` double · `p` split · `enter` deal
- **Roulette:** `↑/↓` pick bet · `-`/`+` straight number · `←/→` chip size ·
  `enter` place · `backspace` clear · `space` spin
- **Hold'em:** `f` fold · `c` check/call · `←/→` raise size · `r` raise ·
  `enter` deal next hand
- `m` returns to the menu from any game.

## Architecture

Built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm-style
loop), [Lip Gloss](https://github.com/charmbracelet/lipgloss) (styling), and
[Bubbles](https://github.com/charmbracelet/bubbles). Hand strength is evaluated
with [chehsunliu/poker](https://github.com/chehsunliu/poker); career state lives
in a pure-Go SQLite database ([modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)).

```
main.go                  open the save DB, launch the program
internal/
  app/                   root model: scene routing + bankroll, main menu, registry
  econ/                  Bankroll + play modes (sandbox / career, daily stipend)
  store/                 SQLite persistence (balance + daily-claim, transactional)
  deck/                  standard 52-card deck / shoe
  ui/                    shared Lip Gloss theme, card & chip rendering, widgets
  games/                 Game contract + SettleMsg/QuitToMenu messages
    slots/ blackjack/ roulette/ holdem/
```

**Money is single-sourced.** Games never touch the database. When a wager
resolves a game emits `games.SettleMsg{Delta}`; the root model applies it to the
bankroll and, in career mode, persists immediately in a transaction so a crash
can't desync the balance. Adding a new game is just a new package implementing
the `tea.Model` contract plus one line in `internal/app/registry.go`.

**Animations** run off a `tea.Tick` frame loop (reels and the roulette ball
decelerate as they approach their predetermined result; cards reveal one at a
time). Results are decided up front, so the animation is purely visual and fair.

## Tests

```sh
go test ./...
```

Coverage focuses on correctness that matters: bankroll arithmetic, the
daily-stipend / persistence logic (including across simulated relaunches),
payout math for every game, and poker hand-ranking sanity (a flush beats a
straight; kickers break ties).

## Roadmap

More games (craps, baccarat, video poker), a stats/achievements screen, and
per-hand history — all deferrable behind the existing `Game` interface and the
schema-versioned save file.
