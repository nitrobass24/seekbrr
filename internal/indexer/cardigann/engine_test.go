package cardigann

import (
	"io"
	stdhttp "net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/autobrr/seekbrr/internal/indexer/cardigann/loader"
)

// engineReplay is an offline Doer that serves a saved body for every request and
// records the outgoing requests, so the Execute path is exercised without any
// network. It serves the same body for each call; the search fixtures issue a
// single request (no login Test block, so EnsureLoggedIn runs Login which is a
// no-op for a def with no Login block).
type engineReplay struct {
	body string

	mu       sync.Mutex
	requests []*stdhttp.Request
}

func (r *engineReplay) Do(req *stdhttp.Request) (*stdhttp.Response, error) {
	r.mu.Lock()
	r.requests = append(r.requests, req)
	r.mu.Unlock()
	return &stdhttp.Response{
		StatusCode: stdhttp.StatusOK,
		Header:     stdhttp.Header{},
		Body:       io.NopCloser(strings.NewReader(r.body)),
		Request:    req,
	}, nil
}

// fixedClock returns a deterministic clock for date defaulting, so a date filter
// that omits the year resolves reproducibly.
func fixedClock() func() time.Time {
	return func() time.Time { return time.Date(2023, time.January, 2, 0, 0, 0, 0, time.UTC) }
}

// loadFixtureDef parses a definition fixture from testdata/ via the loader, so the
// test exercises the same parse path production uses (schema validation + the
// order-preserving FieldsBlock decode).
func loadFixtureDef(t *testing.T, name string) *loader.Definition {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading fixture %q: %v", name, err)
	}
	def, err := loader.Parse(data)
	if err != nil {
		t.Fatalf("parsing fixture %q: %v", name, err)
	}
	return def
}

// readBody loads a saved response body fixture.
func readBody(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading body %q: %v", name, err)
	}
	return data
}

// newFixtureEngine builds an Engine for a fixture definition with a fixed clock
// (deterministic dates) and no Doer (offline ParseResponse path only).
func newFixtureEngine(t *testing.T, defName string) *Engine {
	t.Helper()
	def := loadFixtureDef(t, defName)
	eng, err := NewEngine(def, WithClock(fixedClock()))
	if err != nil {
		t.Fatalf("NewEngine(%q): %v", defName, err)
	}
	return eng
}

// TestParseResponse_HTMLScrape replays a saved HTML response end-to-end and
// asserts the normalized releases. Regression snapshot (self-generated, NOT
// Jackett-diffed): Phase 2 swaps these for differential goldens.
func TestParseResponse_HTMLScrape(t *testing.T) {
	t.Parallel()
	eng := newFixtureEngine(t, "html_scrape.yml")

	releases, err := eng.ParseResponse(readBody(t, "html_scrape.html"), "")
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	if len(releases) != 2 {
		t.Fatalf("releases = %d, want 2", len(releases))
	}

	r0 := releases[0]
	assertTitle(t, r0.Title, "Big Buck Bunny 1080p")
	assertSize(t, r0.Size, 2_684_354_560) // 2.5 GiB (Jackett GetBytes is 1024-based)
	assertSeeders(t, r0, 42, 7)
	assertCategories(t, r0.Categories, []int{2000})
	if r0.DownloadVolumeFactor != 0 {
		t.Errorf("dvf = %v, want 0 (freeleech)", r0.DownloadVolumeFactor)
	}
	if r0.PublishDate != "2021-03-04T09:15:00Z" {
		t.Errorf("publishDate = %q, want 2021-03-04T09:15:00Z", r0.PublishDate)
	}

	r1 := releases[1]
	assertCategories(t, r1.Categories, []int{5000})
	if r1.DownloadVolumeFactor != 1 {
		t.Errorf("row1 dvf = %v, want 1 (no freeleech)", r1.DownloadVolumeFactor)
	}
}

// TestParseResponse_JSONAPI replays a JSON-API response and asserts normalized
// releases (multi-category mapping, byte sizes, RFC3339 dates).
func TestParseResponse_JSONAPI(t *testing.T) {
	t.Parallel()
	eng := newFixtureEngine(t, "json_api.yml")

	releases, err := eng.ParseResponse(readBody(t, "json_api.json"), "json")
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	if len(releases) != 2 {
		t.Fatalf("releases = %d, want 2", len(releases))
	}

	assertTitle(t, releases[0].Title, "Tears of Steel 4K")
	assertSize(t, releases[0].Size, 6_442_450_944)
	assertCategories(t, releases[0].Categories, []int{2000})
	if releases[0].Link != "https://json.invalid/dl/201.torrent" {
		t.Errorf("link = %q", releases[0].Link)
	}
	if releases[0].PublishDate != "2022-06-01T10:30:00Z" {
		t.Errorf("publishDate = %q, want 2022-06-01T10:30:00Z", releases[0].PublishDate)
	}
	assertCategories(t, releases[1].Categories, []int{5000})
}

// TestParseResponse_MagnetSynth proves info-hash -> magnet synthesis for a public
// indexer (the normalizer seam): an infohash-only row yields a magnet URI.
func TestParseResponse_MagnetSynth(t *testing.T) {
	t.Parallel()
	eng := newFixtureEngine(t, "magnet_synth.yml")

	releases, err := eng.ParseResponse(readBody(t, "magnet_synth.html"), "")
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	if len(releases) != 1 {
		t.Fatalf("releases = %d, want 1", len(releases))
	}
	r := releases[0]
	if r.InfoHash != "0123456789abcdef0123456789abcdef01234567" {
		t.Errorf("infohash = %q", r.InfoHash)
	}
	wantMagnet := "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567"
	if got := r.Magnet[:len(wantMagnet)]; got != wantMagnet {
		t.Errorf("magnet prefix = %q, want %q", got, wantMagnet)
	}
}

// TestParseResponse_ResultOrder proves the loader field-order fix end-to-end: the
// category field reads an EARLIER field (_raw_cat) via .Result, so processing
// must follow definition order. Row 0 has an explicit kind (cat 1 -> Movies); row
// 1 omits it and falls back to cat 2 (-> TV) via the field template.
func TestParseResponse_ResultOrder(t *testing.T) {
	t.Parallel()
	eng := newFixtureEngine(t, "result_order.yml")

	releases, err := eng.ParseResponse(readBody(t, "result_order.html"), "")
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	if len(releases) != 2 {
		t.Fatalf("releases = %d, want 2", len(releases))
	}
	assertCategories(t, releases[0].Categories, []int{2000}) // explicit cat 1 -> Movies
	assertCategories(t, releases[1].Categories, []int{5000}) // fallback cat 2 -> TV
}

// TestParseResponse_TodayDefault proves the .Today template namespace is wired
// from the injected clock: a row with no date span falls back to the field
// default "{{ .Today.Year }}-01-01", which must render the fixed-clock year
// (2023) rather than empty. This locks the clock->template seam, which is
// otherwise invisible (the clock only reaching dateparse would leave .Today empty).
func TestParseResponse_TodayDefault(t *testing.T) {
	t.Parallel()
	eng := newFixtureEngine(t, "today_default.yml")

	releases, err := eng.ParseResponse(readBody(t, "today_default.html"), "")
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	if len(releases) != 1 {
		t.Fatalf("releases = %d, want 1", len(releases))
	}
	if releases[0].PublishDate != "2023-01-01" {
		t.Errorf("publishDate = %q, want 2023-01-01 (fixed-clock .Today.Year)", releases[0].PublishDate)
	}
}

// TestResultsJSON_Deterministic proves the marshal seam yields stable bytes.
func TestResultsJSON_Deterministic(t *testing.T) {
	t.Parallel()
	eng := newFixtureEngine(t, "html_scrape.yml")

	releases, err := eng.ParseResponse(readBody(t, "html_scrape.html"), "")
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	a, err := eng.ResultsJSON(releases)
	if err != nil {
		t.Fatalf("ResultsJSON: %v", err)
	}
	b, err := eng.ResultsJSON(releases)
	if err != nil {
		t.Fatalf("ResultsJSON: %v", err)
	}
	if string(a) != string(b) {
		t.Errorf("ResultsJSON not deterministic")
	}
}

// TestExecute_OnlineReplay drives the full Search path (EnsureLoggedIn then
// search.Execute) against a replay Doer, proving the online half assembles: a
// request is issued and the served body is parsed into releases. No login block
// means EnsureLoggedIn is a no-op success. No live HTTP occurs.
func TestExecute_OnlineReplay(t *testing.T) {
	t.Parallel()

	def := loadFixtureDef(t, "html_scrape.yml")
	doer := &engineReplay{body: string(readBody(t, "html_scrape.html"))}
	eng, err := NewEngine(def, WithClock(fixedClock()), WithDoer(doer))
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	releases, err := eng.Search(Query{Keywords: "bunny"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(releases) != 1 {
		// andmatch keeps only the row whose title contains "bunny".
		t.Fatalf("releases = %d, want 1 (andmatch on 'bunny')", len(releases))
	}
	assertTitle(t, releases[0].Title, "Big Buck Bunny 1080p")

	if len(doer.requests) == 0 {
		t.Fatal("no request issued")
	}
	if got := doer.requests[len(doer.requests)-1].URL.Query().Get("q"); got != "bunny" {
		t.Errorf("search query q = %q, want bunny", got)
	}
}

func assertTitle(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
}

func assertSize(t *testing.T, got, want int64) {
	t.Helper()
	if got != want {
		t.Errorf("size = %d, want %d", got, want)
	}
}

func assertSeeders(t *testing.T, r *Release, seeders, leechers int64) {
	t.Helper()
	if r.Seeders != seeders {
		t.Errorf("seeders = %d, want %d", r.Seeders, seeders)
	}
	if r.Leechers != leechers {
		t.Errorf("leechers = %d, want %d", r.Leechers, leechers)
	}
}

func assertCategories(t *testing.T, got, want []int) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("categories = %v, want %v", got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("categories = %v, want %v", got, want)
			return
		}
	}
}
