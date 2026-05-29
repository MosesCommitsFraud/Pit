// Command pit is a terminal casino.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"pit/internal/app"
	"pit/internal/store"
)

func main() {
	// Career persistence is best-effort: if the database can't be opened we
	// still run, with career mode falling back to a fresh bankroll.
	var st *store.Store
	if path, err := store.DefaultPath(); err == nil {
		if s, err := store.Open(path); err == nil {
			st = s
			defer st.Close()
		} else {
			fmt.Fprintln(os.Stderr, "warning: career save disabled:", err)
		}
	}

	root := app.New(st, app.Registry())
	p := tea.NewProgram(root, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "pit:", err)
		os.Exit(1)
	}
}
