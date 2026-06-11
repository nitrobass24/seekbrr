package filter

import (
	"fmt"
	"regexp"
	"unicode"
)

// NOTE: regex filters use RE2 (stdlib regexp) directly for now. Phase 1 item 7
// (regexadapter) will route BOTH these filters and the template's regex through
// a shared .NET-aware adapter (regexp2 on opt-in / non-Latin / RE2-incompatible
// patterns). Do not add regexp2 routing here — that is item 7's seam.

// filterReReplace implements re_replace[pattern,repl]: regex replace-all with
// $1-style backreferences. Jackett template-applies the replacement first; that
// is item 7's concern, so we treat it as a literal Go replacement template.
func filterReReplace(value string, args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("re_replace needs 2 args, got %d: %w", len(args), errMissingArg)
	}
	re, err := regexp.Compile(args[0])
	if err != nil {
		return "", fmt.Errorf("re_replace: compiling pattern: %w", err)
	}
	return re.ReplaceAllString(value, args[1]), nil
}

// filterRegexp implements regexp[pattern]: match the pattern and return capture
// group 1's value. Jackett returns Match.Groups[1].Value, which is "" when the
// pattern does not match or has no group 1 — so a no-match yields "".
func filterRegexp(value string, args []string) (string, error) {
	pattern := firstArg(args)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("regexp: compiling pattern: %w", err)
	}
	m := re.FindStringSubmatch(value)
	if len(m) < 2 {
		// No match, or a match with no capture group 1: Jackett returns "".
		return "", nil
	}
	return m[1], nil
}

// isNonSpacingMark reports whether r is a Unicode non-spacing mark (category
// Mn), matching .NET's UnicodeCategory.NonSpacingMark used by the diacritics
// filter.
func isNonSpacingMark(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}
