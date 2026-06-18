package currency

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var rupiahPrinter = message.NewPrinter(language.Indonesian)

// FormatRupiah formats an integer amount as an Indonesian Rupiah string
// (e.g. 910120000000 → "Rp910.120.000.000").
func FormatRupiah(amount int64) string {
	return rupiahPrinter.Sprintf("Rp%d", amount)
}

// FormatRupiahChange formats a signed amount with a leading "+" for surplus
// or "-" for deficit (e.g. 34200000 → "+Rp34.200.000", -15800000 → "-Rp15.800.000").
func FormatRupiahChange(amount int64) string {
	if amount < 0 {
		return rupiahPrinter.Sprintf("-Rp%d", -amount)
	}
	return rupiahPrinter.Sprintf("+Rp%d", amount)
}
