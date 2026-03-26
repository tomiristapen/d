package product

import "strings"

func normalizeBarcode(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	var digits strings.Builder
	digits.Grow(len(raw))
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	return digits.String()
}

func isValidBarcode(barcode string) bool {
	if barcode == "" {
		return false
	}
	if len(barcode) < 8 || len(barcode) > 14 {
		return false
	}
	for _, r := range barcode {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func barcodeLookupCandidates(raw string) []string {
	barcode := normalizeBarcode(raw)
	if !isValidBarcode(barcode) {
		return nil
	}

	out := []string{barcode}
	if len(barcode) == 12 {
		out = append(out, "0"+barcode)
	}
	if len(barcode) == 13 && strings.HasPrefix(barcode, "0") {
		out = append(out, barcode[1:])
	}
	return out
}

func canonicalBarcode(raw string, fallback string) string {
	if barcode := normalizeBarcode(raw); isValidBarcode(barcode) {
		return barcode
	}
	return normalizeBarcode(fallback)
}
