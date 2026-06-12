package filter

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Row filters operate on the row SET (RowsBlock.Filters), not on a single field
// value, so they are exposed as reusable helpers rather than chained through
// Apply. Their APPLICATION to the parsed row set is wired by the selector
// stage (item 5) and the end-to-end Definition walk (item 10). The registry
// only needs to KNOW their names (see RowFilterKnown) so the corpus
// completeness test sees zero unknown filters.

// andMatchSplit mirrors Jackett's MatchQueryStringAND tokenizer: split on runs
// of non-word characters (.NET Regex "[^\\w]+"). RE2's bare \w is ASCII-only, so
// it would treat a whole Cyrillic/Chinese keyword as one non-word run and drop
// every token, silently disabling the AND match for non-Latin queries. .NET's \w
// is Unicode-aware (\p{L}\p{Mn}\p{Nd}\p{Pc}); we spell that out so tokenization
// matches Jackett for non-Latin keywords. For ASCII this equals [A-Za-z0-9_].
var andMatchSplit = regexp.MustCompile(`[^\p{L}\p{Mn}\p{Nd}\p{Pc}]+`)

// andMatchStopwords are the short words Jackett drops from the keyword set
// before requiring an AND-match.
var andMatchStopwords = map[string]struct{}{"and": {}, "the": {}, "an": {}}

// AndMatch implements the core andmatch row test: tokenize keywords on
// non-word runs, drop tokens of length ≤1 and the stopwords, and keep the row
// iff its title contains every remaining token (case-insensitively). Jackett
// additionally skips this filter entirely for ID-based searches (imdb/tmdb/…)
// and supports an optional character-limit arg on the keywords — that
// search-context gating is the caller's job (items 5/10); this helper covers
// the title-vs-keywords matching itself.
//
// NOTE: RowFilterBlock.Args is intentionally NOT threaded here yet. Row-filter
// application (and any optional andmatch arg) is wired by the selector stage
// (item 5) and the end-to-end walk (item 10); the signature will gain the arg
// at that seam if the corpus requires it. The completeness gate checks only the
// filter NAME, not its arg shape.
func AndMatch(title, keywords string) bool {
	lowerTitle := strings.ToLower(title)
	for _, raw := range andMatchSplit.Split(keywords, -1) {
		tok := strings.ToLower(raw)
		// Jackett drops tokens of length ≤1 by CHARACTER count (.NET string.Length),
		// so count runes, not bytes, or a single Cyrillic/CJK token (2–3 bytes)
		// would survive where Jackett discards it.
		if utf8.RuneCountInString(tok) <= 1 {
			continue
		}
		if _, stop := andMatchStopwords[tok]; stop {
			continue
		}
		if !strings.Contains(lowerTitle, tok) {
			return false
		}
	}
	return true
}

// StrDump implements the strdump row filter: Jackett only debug-logs the row
// and keeps it, so this is a passthrough that always retains the row.
func StrDump(_ string) bool {
	return true
}

// rowFilterNames is the bounded set of row-filter names from the schema
// vocabulary. They are recognized (so the corpus test passes) but applied by
// items 5/10, not by Apply.
var rowFilterNames = map[string]struct{}{
	"andmatch": {},
	"strdump":  {},
}

// RowFilterKnown reports whether name is a recognized row filter.
func RowFilterKnown(name string) bool {
	_, ok := rowFilterNames[name]
	return ok
}
