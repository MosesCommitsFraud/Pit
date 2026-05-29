// Package store persists the career bankroll in a local SQLite database.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store wraps the SQLite connection holding career state.
type Store struct {
	db *sql.DB
}

// State is the persisted career snapshot.
type State struct {
	Balance   int64
	LastClaim string // local date "2006-01-02", empty if never claimed
}

// DefaultPath returns ~/.config/pit/save.db.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "pit", "save.db"), nil
}

// Open opens (creating if needed) the database at path and runs migrations.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // single-writer local file
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS player (
			id              INTEGER PRIMARY KEY CHECK (id = 1),
			balance         INTEGER NOT NULL,
			last_claim_date TEXT NOT NULL DEFAULT ''
		);
		INSERT OR IGNORE INTO player (id, balance, last_claim_date) VALUES (1, 0, '');
	`)
	return err
}

// Load reads the current career state, ensuring the singleton row exists.
func (s *Store) Load() (State, error) {
	var st State
	err := s.db.QueryRow(`SELECT balance, last_claim_date FROM player WHERE id = 1`).
		Scan(&st.Balance, &st.LastClaim)
	return st, err
}

// SaveBalance writes the balance back atomically.
func (s *Store) SaveBalance(balance int64) error {
	_, err := s.db.Exec(`UPDATE player SET balance = ? WHERE id = 1`, balance)
	return err
}

// ClaimDaily grants stipend once per local day. If lastClaim already equals
// today the balance is unchanged and claimed is false. It returns the resulting
// balance and whether a stipend was granted, in a single transaction.
func (s *Store) ClaimDaily(today string, stipend int64) (balance int64, claimed bool, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback()

	var bal int64
	var last string
	if err = tx.QueryRow(`SELECT balance, last_claim_date FROM player WHERE id = 1`).
		Scan(&bal, &last); err != nil {
		return 0, false, err
	}

	if last != today {
		bal += stipend
		claimed = true
		if _, err = tx.Exec(
			`UPDATE player SET balance = ?, last_claim_date = ? WHERE id = 1`,
			bal, today,
		); err != nil {
			return 0, false, err
		}
	}
	if err = tx.Commit(); err != nil {
		return 0, false, err
	}
	return bal, claimed, nil
}
