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

func TrimAbbreviation(s string) (name string, abbr string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ""
	}

	// Normalize spacing
	s = strings.Join(strings.Fields(s), " ")

	// Known abbreviation tokens (single words)
	abbrTokens := map[string]struct{}{
		"MR": {}, "MR.": {},
		"MISS": {},
		"MS":   {},
		"MRS":  {},
		"DR":   {}, "DR.": {},
		"PROF": {}, "PROF.": {}, "PROFESSOR": {},
		"DATO": {}, "DATO'": {},
		"DATUK":   {},
		"DATIN":   {},
		"TAN SRI": {}, "TAN SERI": {},
		"SRI": {}, "SERI": {},
		"PUAN": {}, "ENCIK": {}, "TUAN": {},
		"ABANG":   {},
		"SENATOR": {},
		"YB":      {}, "YB.": {}, "Y.B.": {}, "B.": {},
		"Y": {}, "Y.": {},
		"YBHG": {}, "BHG": {}, "BHG.": {},
		"YBM": {}, "YBM.": {}, "Y.B.M.": {},
		"YDH": {}, "YDH.": {}, "Y.D.H.": {}, "DH.": {},
	}

	words := strings.Split(s, " ")

	// Walk from right → left, building suffix
	var suffix []string
	var names []string

	for i := len(words) - 1; i >= 0; i-- {
		if _, ok := abbrTokens[words[i]]; !ok {
			names = append([]string{words[i]}, names...)
			continue
		}

		suffix = append([]string{words[i]}, suffix...)
	}

	name = strings.TrimSpace(strings.Join(names, " "))
	abbr = strings.TrimSpace(strings.Join(suffix, " "))

	return name, abbr
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
