package utils

import (
	"strings"
)

var atomicTitles = map[string]bool{
	"YABHG": true,
	"YBHG":  true,
	"YAB":   true,
	"YBM":   true,
	"YTM":   true,
	"YAM":   true,
	"YM":    true,
	"YB":    true,

	"SENATOR": true,

	"TUN":   true,
	"DATO":  true,
	"DATO'": true,
	"DATUK": true,
	"DATIN": true,
	"PUAN":  true,
	"ENCIK": true,
	"CIK":   true,
	"DR":    true,
	"IR":    true,
	"HAJI":  true,
	"MR":    true,
	"MISS":  true,
	"MADAM": true,
}

var compoundTitles = map[string]bool{
	"TAN SRI":    true,
	"TAN SERI":   true,
	"DATO' SRI":  true,
	"DATO SERI":  true,
	"DATUK SERI": true,
	"Y B":        true,
	"Y B DATO":   true,
	"Y B DATO'":  true,
	"Y B DATUK":  true,
}

func SplitTitle(fullName string) (title string, name string) {
	n := normalize(fullName)
	tokens := strings.Fields(n)

	var consumed []string
	i := 0

	for i < len(tokens) {
		// Try compound titles first
		if i+1 < len(tokens) {
			pair := tokens[i] + " " + tokens[i+1]
			if compoundTitles[pair] {
				consumed = append(consumed, pair)
				i += 2
				continue
			}
		}

		// Single-token titles
		if atomicTitles[tokens[i]] {
			consumed = append(consumed, tokens[i])
			i++
			continue
		}

		break
	}

	title = strings.Join(consumed, " ")
	name = strings.Join(tokens[i:], " ")
	return title, name
}

func normalize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	s = strings.ReplaceAll(s, "’", "'")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, ".", "")

	s = strings.ReplaceAll(s, "Y B", "YB")
	s = strings.ReplaceAll(s, " B ", " BIN ")
	s = strings.ReplaceAll(s, " HJ ", " HAJI ")
	s = strings.ReplaceAll(s, " HJH ", " HAJAH ")

	// collapse whitespace
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func normalizeName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	// collapse whitespace
	s = strings.Join(strings.Fields(s), " ")
	return s
}
