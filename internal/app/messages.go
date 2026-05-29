package app

import "pit/internal/econ"

// chooseModeMsg is emitted by the menu when the player picks a play mode.
type chooseModeMsg struct{ mode econ.Mode }

// startGameMsg is emitted by the menu when the player launches a game.
type startGameMsg struct{ id string }
