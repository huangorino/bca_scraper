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

// StringValue returns the value of a string pointer or an empty string if nil.
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// TimeValue returns the value of a time.Time pointer or a zero time if nil.
func TimeValue(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// PtrString returns a pointer to the given string.
func PtrString(s string) *string {
	return &s
}

// IntValue returns the value of an int pointer or 0 if nil.
func IntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
