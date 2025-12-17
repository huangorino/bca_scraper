package utils

import (
	"fmt"
	"os"
	"regexp"
	"sort"
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

var canonicalTitles = []string{
	"SENATOR DATO'",
	"SENATOR DATUK",
	"SENATOR",
	"YABHG",
	"YBHG",
	"YAB",
	"YBM",
	"YTM",
	"YAM",
	"YM",
	"YB",
	"TUN",
	"TAN SRI DATO'",
	"TAN SRI",
	"DATO' SRI",
	"DATO'",
	"DATO",
	"DATUK SERI",
	"DATUK",
	"DATIN",
	"PUAN",
	"ENCIK",
	"CIK",
	"DR",
	"IR",
	"HAJI",
	"MR",
	"MISS",
	"MADAM",
}

var titleRegex []*regexp.Regexp

func init() {
	// Sort by word count DESC, then length DESC
	sort.SliceStable(canonicalTitles, func(i, j int) bool {
		wi := len(strings.Fields(canonicalTitles[i]))
		wj := len(strings.Fields(canonicalTitles[j]))
		if wi != wj {
			return wi > wj
		}
		return len(canonicalTitles[i]) > len(canonicalTitles[j])
	})

	for _, t := range canonicalTitles {
		pattern := fmt.Sprintf(`^(%s)(\s+|$)`, regexp.QuoteMeta(t))
		titleRegex = append(titleRegex, regexp.MustCompile(pattern))
	}
}

func SplitTitle(fullName string) (title string, name string) {
	n := normalize(fullName)

	for _, re := range titleRegex {
		if re.MatchString(n) {
			m := re.FindStringSubmatch(n)
			title = m[1]
			name = strings.TrimSpace(n[len(m[0]):])
			return title, name
		}
	}

	return "", n
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
