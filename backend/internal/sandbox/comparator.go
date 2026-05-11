package sandbox

import "strings"

// Equal compares the expected output of a test case with whatever the user's
// program actually printed. It's NOT a strict string equals — we forgive some
// stuff that doesn't usually matter:
//   - windows vs unix line endings (\r\n vs \n)
//   - trailing spaces/tabs on each line
//   - extra blank lines at the very end
//
// If you ever need exact match (eg for problems where whitespace matters)
// you'd add a different comparator and let the problem pick which one.
func Equal(expected, actual string) bool {
	return normalize(expected) == normalize(actual)
}

// normalize is the shared cleanup we do before comparing. It tries to make
// two stings that "look the same" actualy be the same.
func normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}
