package torznab

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/autobrr/harbrr/internal/indexer/cardigann/mapper"
	"github.com/autobrr/harbrr/internal/indexer/cardigann/normalizer"
	"github.com/autobrr/harbrr/internal/indexer/cardigann/search"
	tzn "github.com/autobrr/harbrr/internal/torznab"
)

// defaultLimit bounds the served result page; it is the same constant the caps
// document advertises as <limits max>, so the advertised and enforced page sizes
// cannot drift. A request limit above it is clamped down.
const defaultLimit = tzn.LimitsMax

// buildQuery maps the Torznab/Newznab request params to the engine's search
// query. The `cat` newznab ids are resolved to tracker category ids through the
// indexer's capabilities (the request-side mirror of the response-side category
// mapping). The book `title` param maps to BookTitle; `rid` is the TVRage id.
// `publisher` has no engine query field and is intentionally ignored (recorded
// as a known divergence).
func buildQuery(q url.Values, caps *mapper.Capabilities) search.Query {
	query := search.Query{
		Keywords:  q.Get("q"),
		IMDBID:    q.Get("imdbid"),
		TMDBID:    q.Get("tmdbid"),
		TVDBID:    q.Get("tvdbid"),
		TVMazeID:  q.Get("tvmazeid"),
		TraktID:   q.Get("traktid"),
		DoubanID:  q.Get("doubanid"),
		RageID:    q.Get("rid"),
		Season:    q.Get("season"),
		Ep:        q.Get("ep"),
		Year:      q.Get("year"),
		Artist:    q.Get("artist"),
		Album:     q.Get("album"),
		Label:     q.Get("label"),
		Track:     q.Get("track"),
		Author:    q.Get("author"),
		BookTitle: q.Get("title"),
	}
	if ids := parseCatList(q.Get("cat")); len(ids) > 0 {
		query.Categories = caps.MapTorznabCapsToTrackers(ids)
	}
	return query
}

// parseCatList parses a comma-separated newznab category list ("2000,5000"),
// dropping blanks and non-numeric entries.
func parseCatList(cat string) []int {
	if strings.TrimSpace(cat) == "" {
		return nil
	}
	parts := strings.Split(cat, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// paging is the resolved limit/offset window for the served result page.
type paging struct {
	limit  int
	offset int
}

// parsePaging reads limit/offset, clamping limit to [1, defaultLimit] and a
// negative offset to 0. A limit at or above the max stays at the max.
func parsePaging(q url.Values) paging {
	limit := defaultLimit
	if v, err := strconv.Atoi(q.Get("limit")); err == nil && v > 0 && v < defaultLimit {
		limit = v
	}
	offset := 0
	if v, err := strconv.Atoi(q.Get("offset")); err == nil && v > 0 {
		offset = v
	}
	return paging{limit: limit, offset: offset}
}

// apply slices releases to the [offset, offset+limit) window with bounds guards
// so an offset past the end yields an empty (not panicking) page.
func (p paging) apply(releases []*normalizer.Release) []*normalizer.Release {
	if p.offset >= len(releases) {
		return nil
	}
	end := p.offset + p.limit
	if end > len(releases) {
		end = len(releases)
	}
	return releases[p.offset:end]
}
