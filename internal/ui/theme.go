// Package ui holds shared styling and rendering helpers for every screen.
//
// The look is "retro high-contrast, mono + crimson": a near-monochrome
// grayscale with a single crimson accent used only for wins, highlights, and
// the red card suits. Layout is top-aligned and full-width, framed by double
// rules, with blocky inverted title chips and uppercase labels.
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette — grayscale plus one accent.
var (
	Accent = lipgloss.Color("#e0242f") // crimson: wins, highlights, red suits
	Bright = lipgloss.Color("#f2f2f2") // primary text
	Text   = lipgloss.Color("#c4c4c4") // normal text
	Dim    = lipgloss.Color("#808080") // secondary text
	Faint  = lipgloss.Color("#4a4a4a") // rules, placeholders
	Black  = lipgloss.Color("#0c0c0c") // text on inverted chips
)

// Shared styles.
var (
	// Chip is an inverted blocky label (bright background, dark text).
	Chip = lipgloss.NewStyle().Bold(true).Foreground(Black).Background(Bright).Padding(0, 1)

	Title = lipgloss.NewStyle().Bold(true).Foreground(Bright)

	Subtle = lipgloss.NewStyle().Foreground(Dim)

	Heading = lipgloss.NewStyle().Bold(true).Foreground(Bright)

	// Label is an uppercase section heading.
	Label = lipgloss.NewStyle().Bold(true).Foreground(Dim)

	AccentText = lipgloss.NewStyle().Bold(true).Foreground(Accent)

	// Selected is the inverted highlight for the focused list item.
	Selected = lipgloss.NewStyle().Bold(true).Foreground(Black).Background(Accent).Padding(0, 1)

	Unselected = lipgloss.NewStyle().Foreground(Text).Padding(0, 1)

	WinText = lipgloss.NewStyle().Bold(true).Foreground(Accent)

	LoseText = lipgloss.NewStyle().Foreground(Dim)

	RuleStyle = lipgloss.NewStyle().Foreground(Faint)

	HelpBar = lipgloss.NewStyle().Foreground(Dim)
)

// Caps uppercases a string for labels/keys.
func Caps(s string) string { return strings.ToUpper(s) }

// Money formats a chip amount with thousands separators and a leading $.
func Money(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := ""
	for n >= 1000 {
		s = "," + pad3(n%1000) + s
		n /= 1000
	}
	s = itoa(n) + s
	if neg {
		return "-$" + s
	}
	return "$" + s
}

func pad3(n int64) string {
	d := itoa(n)
	for len(d) < 3 {
		d = "0" + d
	}
	return d
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
