package utils

import (
	"os"
	"strings"
	"time"
)

// CleanString removes unwanted characters, spaces, and formatting from raw text.
// - Replaces colons
// - Removes non-breaking spaces (\u00a0)
// - Trims whitespace
func CleanString(s string) string {
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	return strings.TrimSpace(s)
}

// ParseDate attempts to normalize date formats like "16 Oct 2025"
// into "YYYY-MM-DD". If parsing fails, returns the raw string.
func ParseDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	layouts := []string{
		"02 Jan 2006",
		"2 Jan 2006",
		"02/Jan/2006",
		"2/Jan/2006",
		"02/01/2006",
		"2/1/2006",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

// Truncate safely trims a string to maxLen characters (UTF-8 safe).
// Useful when logging or inserting into DB columns with length limits.
func Truncate(s string, maxLen int) string {
	if len([]rune(s)) <= maxLen {
		return s
	}
	return string([]rune(s)[:maxLen])
}

// IsEmpty checks if a string is empty or whitespace.
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// htmlUnescape replaces HTML encoded characters like &amp; with &
func HtmlUnescape(s string) string {
	return strings.ReplaceAll(s, "&amp;", "&")
}

// SaveToFile writes the given content to a file at the specified path.
func SaveToFile(filePath string, content []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return err
	}

	return nil
}

// StripMarkdown removes markdown code blocks from a string.
func StripMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if after, ok := strings.CutPrefix(s, "```json"); ok {
		s = after
	}
	if after, ok := strings.CutPrefix(s, "```"); ok {
		s = after
	}
	if after, ok := strings.CutSuffix(s, "```"); ok {
		s = after
	}
	return s
}

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

	// normalize unicode apostrophe
	s = strings.ReplaceAll(s, "’", "'")

	// remove dots (YABHG., DR., etc.)
	s = strings.ReplaceAll(s, ".", "")

	// collapse whitespace
	s = strings.Join(strings.Fields(s), " ")
	return s
}
