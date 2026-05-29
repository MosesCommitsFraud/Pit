// Package ui holds shared styling and rendering helpers for every screen.
package ui

import "github.com/charmbracelet/lipgloss"

// Palette colors used across the casino.
var (
	Felt   = lipgloss.Color("#0b6b3a") // table green
	Gold   = lipgloss.Color("#e8c14a") // chips / accents
	Cream  = lipgloss.Color("#f3ead3") // primary text
	Muted  = lipgloss.Color("#8a9a8f") // secondary text
	Red    = lipgloss.Color("#d6453d") // losses / red suits
	Green  = lipgloss.Color("#5fd07a") // wins
	Ink    = lipgloss.Color("#10130f") // dark text on light cards
	Border = lipgloss.Color("#1f7a48")
)

// Shared styles.
var (
	Title = lipgloss.NewStyle().Bold(true).Foreground(Gold)

	Subtle = lipgloss.NewStyle().Foreground(Muted)

	Heading = lipgloss.NewStyle().Bold(true).Foreground(Cream)

	Selected = lipgloss.NewStyle().Bold(true).Foreground(Ink).Background(Gold).Padding(0, 1)

	Unselected = lipgloss.NewStyle().Foreground(Cream).Padding(0, 1)

	WinText = lipgloss.NewStyle().Bold(true).Foreground(Green)

	LoseText = lipgloss.NewStyle().Bold(true).Foreground(Red)

	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(1, 2)

	HelpBar = lipgloss.NewStyle().Foreground(Muted)
)

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
