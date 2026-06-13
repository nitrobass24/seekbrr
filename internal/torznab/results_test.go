package torznab

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/autobrr/harbrr/internal/indexer/cardigann/normalizer"
)

// fixedNow is the deterministic pubDate fallback clock for the results goldens.
func fixedNow() time.Time { return time.Date(2026, time.June, 13, 12, 0, 0, 0, time.UTC) }

func demoFeed() FeedInfo {
	return FeedInfo{
		IndexerID:   "demo",
		Name:        "Demo Tracker",
		Description: "Synthetic tracker for the Torznab results goldens.",
		SiteLink:    "https://demo.test/",
		Type:        "public",
		SelfURL:     "https://harbrr.local/api/v2.0/indexers/demo/results/torznab",
	}
}

// fullRelease exercises every emitted field: standard + custom categories, a
// download link carrying a (synthetic) passkey — intended served output — a
// freeleech downloadvolumefactor of 0, external ids, media fields, and a dated
// release with a non-UTC offset (to pin the RFC1123Z rendering).
func fullRelease() *normalizer.Release {
	return &normalizer.Release{
		Title:                "Example.Movie.2024.1080p.BluRay",
		Details:              "https://demo.test/torrent/1",
		Link:                 "https://demo.test/download/1.torrent?passkey=synthetic-demo-key",
		Size:                 5368709120,
		Categories:           []int{2040, 100001},
		Seeders:              12,
		Leechers:             3,
		Peers:                15,
		Grabs:                7,
		Files:                4,
		PublishDate:          "2024-03-14T17:10:42-04:00",
		DownloadVolumeFactor: 0, // freeleech: 0 must be emitted, not dropped
		UploadVolumeFactor:   1,
		MinimumRatio:         1.5,
		MinimumSeedTime:      172800,
		IMDBID:               "tt0903747",
		TMDBID:               1396,
		TVDBID:               81189,
		Year:                 2024,
		Genre:                "Drama,Crime",
		Poster:               "https://demo.test/poster/1.jpg",
	}
}

// magnetOnlyRelease has no download link: guid/link/enclosure all fall back to
// the magnet (which carries an & that must XML-escape to &amp;), and size is 0
// (must render <size>0</size> and length="0").
func magnetOnlyRelease() *normalizer.Release {
	return &normalizer.Release{
		Title:                "Magnet Only Release",
		Magnet:               "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567&dn=Magnet+Only+Release",
		InfoHash:             "0123456789ABCDEF0123456789ABCDEF01234567",
		Size:                 0,
		Categories:           []int{5070},
		Seeders:              0,
		Peers:                0,
		DownloadVolumeFactor: 1,
		UploadVolumeFactor:   1,
	}
}

// minimalBadCharRelease has only the required fields, no date (pubDate falls
// back to now), and a control char (0x1A) in the title that must be stripped.
func minimalBadCharRelease() *normalizer.Release {
	return &normalizer.Release{
		Title:                "Minimal\x1aRelease",
		Link:                 "https://demo.test/download/2.torrent",
		Size:                 1048576,
		Categories:           []int{2030},
		Seeders:              1,
		Peers:                1,
		DownloadVolumeFactor: 1,
		UploadVolumeFactor:   1,
	}
}

func TestMarshalResultsGolden(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		golden   string
		releases []*normalizer.Release
	}{
		{
			name:     "feed",
			golden:   "results/feed.xml",
			releases: []*normalizer.Release{fullRelease(), magnetOnlyRelease(), minimalBadCharRelease()},
		},
		{
			name:     "empty",
			golden:   "results/empty.xml",
			releases: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := MarshalResults(demoFeed(), tt.releases, fixedNow())
			if err != nil {
				t.Fatalf("MarshalResults: %v", err)
			}
			assertGolden(t, tt.golden, got)
			assertWellFormed(t, got)
		})
	}
}

// TestResultsGuidPrecedence pins Jackett's FixResults guid precedence:
// Link, else Magnet, else Details.
func TestResultsGuidPrecedence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		r    *normalizer.Release
		want string
	}{
		{"link wins", &normalizer.Release{Link: "L", Magnet: "M", Details: "D"}, "L"},
		{"magnet when no link", &normalizer.Release{Magnet: "M", Details: "D"}, "M"},
		{"details last", &normalizer.Release{Details: "D"}, "D"},
	}
	for _, tt := range tests {
		if got := GUIDFor(tt.r); got != tt.want {
			t.Errorf("%s: GUIDFor = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestResultsZeroSizeAndFreeleech confirms <size>0</size>, enclosure length="0",
// and a freeleech downloadvolumefactor of 0 are all emitted (not dropped).
func TestResultsZeroSizeAndFreeleech(t *testing.T) {
	t.Parallel()
	got, err := MarshalResults(demoFeed(), []*normalizer.Release{magnetOnlyRelease()}, fixedNow())
	if err != nil {
		t.Fatalf("MarshalResults: %v", err)
	}
	s := string(got)
	for _, want := range []string{
		"<size>0</size>",
		`length="0"`,
		`name="seeders" value="0"`,
		`name="downloadvolumefactor" value="1"`,
		"&amp;dn=Magnet", // the magnet & is XML-escaped, not mangled
	} {
		if !strings.Contains(s, want) {
			t.Errorf("results missing %q in:\n%s", want, s)
		}
	}
}

// TestResultsStripsInvalidXMLChars confirms a control char in a title is removed
// before marshaling (parity with Jackett's RemoveInvalidXMLChars).
func TestResultsStripsInvalidXMLChars(t *testing.T) {
	t.Parallel()
	got, err := MarshalResults(demoFeed(), []*normalizer.Release{minimalBadCharRelease()}, fixedNow())
	if err != nil {
		t.Fatalf("MarshalResults: %v", err)
	}
	s := string(got)
	if strings.ContainsRune(s, 0x1A) {
		t.Error("results contain the raw 0x1A control char")
	}
	if !strings.Contains(s, "<title>MinimalRelease</title>") {
		t.Errorf("title not sanitized as expected:\n%s", s)
	}
}

// TestResultsEmptyFeedHasChannel confirms a no-results feed is a valid feed with
// a full <channel> header and zero items, not a bare/empty document.
func TestResultsEmptyFeedHasChannel(t *testing.T) {
	t.Parallel()
	got, err := MarshalResults(demoFeed(), nil, fixedNow())
	if err != nil {
		t.Fatalf("MarshalResults: %v", err)
	}
	s := string(got)
	for _, want := range []string{"<channel>", "<title>Demo Tracker</title>", "<language>en-US</language>"} {
		if !strings.Contains(s, want) {
			t.Errorf("empty feed missing %q", want)
		}
	}
	if strings.Contains(s, "<item>") {
		t.Error("empty feed should contain no <item>")
	}
}

// TestSanitizeXMLText pins the precise strip set: control chars, BOM and the
// U+FFFE/U+FFFF non-characters and invalid UTF-8 bytes are removed; tab/newline/
// CR and a genuine (3-byte) U+FFFD REPLACEMENT CHARACTER are preserved (Jackett's
// regex strips lone surrogates, not a well-formed U+FFFD).
func TestSanitizeXMLText(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"control char", "Bad\x1aChar", "BadChar"},
		{"tab/newline/cr preserved", "a\tb\nc\rd", "a\tb\nc\rd"},
		{"bom stripped", "\uFEFFhello", "hello"},
		{"genuine U+FFFD preserved", "a\uFFFDb", "a\uFFFDb"},
		{"invalid byte stripped", "a\x80b", "ab"},
		{"noncharacter stripped", "a\uFFFFb", "ab"},
		{"clean string untouched", "Normal Title 2024", "Normal Title 2024"},
		{"astral preserved", "emoji \U0001F600 ok", "emoji \U0001F600 ok"},
	}
	for _, tt := range tests {
		if got := sanitizeXMLText(tt.in); got != tt.want {
			t.Errorf("%s: sanitizeXMLText(%q) = %q, want %q", tt.name, tt.in, got, tt.want)
		}
	}
}

// TestResultsFutureDateClamp confirms a release dated after now is clamped to now
// in pubDate (Jackett's FixResults future-date clamp).
func TestResultsFutureDateClamp(t *testing.T) {
	t.Parallel()
	future := &normalizer.Release{
		Title: "Future Dated", Link: "https://demo.test/f.torrent", Size: 1,
		Categories: []int{2000}, Seeders: 1, Peers: 1,
		PublishDate:          "2099-01-01T00:00:00Z",
		DownloadVolumeFactor: 1, UploadVolumeFactor: 1,
	}
	got, err := MarshalResults(demoFeed(), []*normalizer.Release{future}, fixedNow())
	if err != nil {
		t.Fatalf("MarshalResults: %v", err)
	}
	wantPub := "<pubDate>" + fixedNow().Format(time.RFC1123Z) + "</pubDate>"
	if !strings.Contains(string(got), wantPub) {
		t.Errorf("future pubDate not clamped to now; want %q in:\n%s", wantPub, got)
	}
}

// TestResultsGenreWireForm confirms the genre attr uses the ", " (comma+space)
// wire join Jackett's ResultPage emits, not harbrr's internal "," form.
func TestResultsGenreWireForm(t *testing.T) {
	t.Parallel()
	r := &normalizer.Release{
		Title: "G", Link: "https://demo.test/g.torrent", Size: 1,
		Categories: []int{2000}, Seeders: 1, Peers: 1, Genre: "Drama,Crime,Thriller",
		DownloadVolumeFactor: 1, UploadVolumeFactor: 1,
	}
	got, err := MarshalResults(demoFeed(), []*normalizer.Release{r}, fixedNow())
	if err != nil {
		t.Fatalf("MarshalResults: %v", err)
	}
	if !strings.Contains(string(got), `name="genre" value="Drama, Crime, Thriller"`) {
		t.Errorf("genre not joined with comma+space:\n%s", got)
	}
}

// assertWellFormed confirms the bytes parse as XML (no malformed output).
func assertWellFormed(t *testing.T, b []byte) {
	t.Helper()
	dec := xml.NewDecoder(strings.NewReader(string(b)))
	for {
		_, err := dec.Token()
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			t.Fatalf("not well-formed XML: %v", err)
		}
	}
}
