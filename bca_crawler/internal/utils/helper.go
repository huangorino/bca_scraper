package utils

import (
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
func ParseDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	layout := "02 Jan 2006"
	if t, err := time.Parse(layout, s); err == nil {
		return t.Format("2006-01-02")
	}
	return s
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
