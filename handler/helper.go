package handler

import (
	"context"
	"strings"
	"unicode"

	log "MgApplication/api-log"

	"golang.org/x/text/unicode/norm"
)

// // stringToUint64 is a helper function to convert a string to uint64
// func stringToUint64(str string) (uint64, error) {
// 	num, err := strconv.ParseUint(str, 10, 64)

// 	return num, err
// }

// // toMap is a helper function to add meta and data to a map
// func toMap(m meta, data any, key string) map[string]any {
// 	return map[string]any{
// 		"meta": m,
// 		key:    data,
// 	}
// }

// ****Normalisation functions****//
// NormalizeAndClean - Normalizes Unicode, replaces special characters, and cleans up spaces.
func NormalizeAndClean(input string) string {
	// Step 1: Normalize to NFC form for consistent representation
	normalized := norm.NFC.String(input)

	// Step 2: Replace Unicode equivalents (dashes, quotes, ellipsis, currencies, zero-width)
	normalized = replaceUnicodeEquivalents(normalized)

	// Step 3: Remove control characters (invisible and non-printable)
	normalized = removeControlChars(normalized)

	// Step 4: Collapse multiple spaces (handles all Unicode whitespaces)
	normalized = collapseSpaces(normalized)

	// Step 5: Trim leading and trailing spaces
	return strings.TrimSpace(normalized)
}

// replaceUnicodeEquivalents - Handles special Unicode replacements.
func replaceUnicodeEquivalents(input string) string {
	return strings.Map(func(r rune) rune {
		// Convert all Unicode spaces (including non-breaking) to a regular ASCII space
		if unicode.IsSpace(r) {
			return ' '
		}

		switch r {
		// Dashes → ASCII hyphen
		case '–', '—', '‒', '−':
			return '-'

		// Single quotes → Apostrophe
		case '‘', '’', '′':
			return '\''

		// Double quotes → Quotation mark
		case '“', '”', '″':
			return '"'

		// Ellipsis → Period
		case '…':
			return '.'

		// Zero-width spaces (remove them)
		case '\u200B', '\u200C', '\u200D', '\uFEFF':
			return -1

		default:
			return r
		}
	}, input)
}

// removeControlChars - Removes control characters (ASCII and Unicode Cc category).
func removeControlChars(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, input)
}

// Collapse multiple spaces (including Unicode spaces)
func collapseSpaces(input string) string {
	var builder strings.Builder
	spaceFound := false

	for _, r := range input {
		if unicode.IsSpace(r) {
			if !spaceFound {
				builder.WriteRune(' ')
				spaceFound = true
			}
		} else {
			builder.WriteRune(r)
			spaceFound = false
		}
	}
	return builder.String()
}

// Faster version of NormalizeAndClean
// NormalizeAndClean - Normalizes Unicode, replaces special characters, and cleans up spaces.
func NormalizeAndClean2(input string) string {
	// Quick scan to check if normalization and cleaning are needed
	if isClean(input) {
		return input // Fast path for already clean inputs
	}

	// Step 1: Normalize to NFC form for consistent representation
	normalized := norm.NFC.String(input)

	// Step 2: Replace Unicode equivalents, remove control characters, and collapse spaces
	var builder strings.Builder
	builder.Grow(len(normalized)) // Pre-allocate capacity

	spaceFound := false

	for _, r := range normalized {
		// Handle control characters and zero-width spaces (skip them)
		if unicode.IsControl(r) || isZeroWidth(r) {
			continue
		}

		// Ignore backslashes (to prevent escape sequences from showing up)
		if r == '\\' {
			continue
		}

		// Convert all Unicode spaces (including non-breaking) to a regular ASCII space
		if unicode.IsSpace(r) {
			if !spaceFound {
				builder.WriteRune(' ')
				spaceFound = true
			}
			continue
		}

		// Handle special Unicode replacements (dashes, quotes, ellipsis, currency symbols)
		r = replaceSpecialRune(r)

		builder.WriteRune(r)
		spaceFound = false
	}
	// Return the cleaned string with leading and trailing spaces trimmed
	normalizedCleantext := strings.TrimSpace(builder.String())
	ctx := context.Background()
	log.Debug(ctx, "Normalized Cleantext: %s", string(normalizedCleantext))
	// fmt.Println(ctx, "Normalized Cleantext:", normalizedCleantext)

	return normalizedCleantext
}

// isZeroWidth - Checks if the rune is a zero-width character.
func isZeroWidth(r rune) bool {
	switch r {
	case '\u200B', '\u200C', '\u200D', '\uFEFF': // Zero-width spaces
		return true
	default:
		return false
	}
}

// replaceSpecialRune - Handles special Unicode replacements.
func replaceSpecialRune(r rune) rune {
	switch r {
	// Dashes → ASCII hyphen
	case '–', '—', '‒', '−':
		return '-'

	// Single quotes → Apostrophe
	case '‘', '’', '′':
		return '\''

	// Double quotes → Quotation mark
	case '“', '”', '″':
		return '"'

	// Ellipsis → Period
	case '…':
		return '.'

	//** Currency symbols (replace with letters)
	// case '₹':
	// 	return 'I'
	// case '€':
	// 	return 'E'
	// case '£':
	// 	return 'L'

	default:
		return r
	}
}

// isClean - Quickly scans the input to check if cleaning is needed.
func isClean(input string) bool {
	spaceFound := false

	for _, r := range input {
		if r > 0x7F { // Non-ASCII character (Unicode check)
			return false
		}
		if unicode.IsControl(r) || isZeroWidth(r) {
			return false
		}
		if unicode.IsSpace(r) {
			if spaceFound {
				return false // Multiple spaces found
			}
			spaceFound = true
		} else {
			spaceFound = false
		}
	}

	return true // No cleaning needed
}

//**End of Normalisation functions**//

// NormalizeAndClean - Normalizes Unicode, replaces special characters, and cleans up spaces.
