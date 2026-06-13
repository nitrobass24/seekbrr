# Torznab serializer fixtures

Golden XML for harbrr's Torznab/Newznab serializer (`internal/torznab`), the
*arr-facing contract. Each golden is harbrr's own canonical, deterministic
output and is byte-compared by the package tests; regenerate with
`go test ./internal/torznab/ -update` only after confirming the output matches
the case's oracle.

## Oracle policy (offline)

Goldens are **not** captured from a live Jackett or a live Sonarr/Radarr (the
project decision; see `../../indexer/cardigann/parity/testdata/README.md`). harbrr
is GPL-2.0, same as Jackett, so porting Jackett's own test material is
license-compatible (`jackett/NOTICE`). Each golden records its `golden_source`:

- **`jackett-port`** — values ported from Jackett's own test assertions
  (`CardigannIndexerTests.TestCardigannTorznabCategories`) or its serializer
  source (`TorznabCapabilities`, `ResultPage`), at the pinned commit
  `b4140c7`. The authoritative offline oracle.
- **`hand-derived`** — values computed by hand from the Torznab/Newznab spec +
  Jackett's serializer semantics + the Sonarr/Radarr request shapes.

## `caps/` — capabilities document (`t=caps`)

| file | golden_source | what it pins |
|------|---------------|--------------|
| `caps/jackett-categories.xml` | jackett-port | The category tree from `TestCardigannTorznabCategories`' 2nd definition: parent/child nesting, custom ids (100044, **137107**, 100045), and the `GetTorznabCategoryTree(true)` sort (standard parents ascending by id, then customs by name). |
| `caps/jackett-modes.xml` | jackett-port | The re-derived `supportedParams` for the 3rd definition: tv-search drops `imdbid` (gated by `AllowTVSearchIMDB`, off here), `audio-search` mirrors `music-search`, all six modes always emitted. |
| `caps/allowrawsearch.xml` | hand-derived | `allowrawsearch` adds `searchEngine="raw"` to every mode; `allowtvsearchimdb: true` makes tv-search advertise `imdbid` (in canonical order `q,season,imdbid`). |

The structural facts behind the `jackett-port` goldens (custom-id hashes, tree
order, supported-param order) are additionally asserted directly in
`../caps_test.go` (`TestCapsCategoryTreeOracle`, `TestCapsSupportedParamsOracle`,
`TestCapsTVImdbGate`), independent of XML whitespace.

## `results/` — results feed (`t=search` and the typed modes)

| file | golden_source | what it pins |
|------|---------------|--------------|
| `results/feed.xml` | hand-derived | The `<item>` element order + `torznab:attr` block from `ResultPage.ToXml`: standard + custom category emission (plain `<category>` and `torznab:attr`), `imdb` (7-digit) vs `imdbid` (`tt`+7-digit), freeleech `downloadvolumefactor=0`, a magnet-only release (guid/link/enclosure fall back to the magnet; `&`→`&amp;`; `<size>0</size>`/`length="0"`), and control-char stripping. |
| `results/empty.xml` | hand-derived | A no-results feed: a valid `<rss>`/`<channel>` with the full header and zero `<item>`s (HTTP 200, never an error). |

The `<item>` grammar (element/attr names + order, RFC1123Z `pubDate`, guid
precedence) is reproduced from `ResultPage.ToXml` / `ReleaseInfo` /
`BaseIndexer.FixResults` at commit `b4140c7`.

## Known divergences from Jackett / the spec

Deliberate or accepted differences, each with an explicit disposition
(`[Tracked: Phase N]` a real gap with a plan follow-up · `[Deliberate]` an
intentional design choice · `[Accepted]` a kept difference, no work planned),
mirroring `../../indexer/cardigann/parity/testdata/README.md`. None is hidden by
a fixture authored to dodge it. The HTTP-handler-specific entries (error-code
status policy, `atom:link` self URL, `cat`/`limit`/`offset`, default categories,
result-category filtering) are added with that commit.

### Caps document

- **`<server title="harbrr">`** — Jackett emits `title="Jackett"`. Cosmetic;
  Sonarr/Radarr ignore it. **`[Deliberate]`**
- **Attribute-only elements render `<e></e>` not `<e/>`** — harbrr uses
  `encoding/xml`, which has no self-closing form; both are well-formed and parse
  identically for *arr. **`[Deliberate]`**
- **Duplicate-custom-id `<category>` collapse** — when two category mappings
  resolve to the same custom id (numeric reuse or a SHA1-uint16 hash collision),
  Jackett's tree carries both `<category>` nodes; harbrr's mapper de-dups
  advertised categories by id (last name wins), so harbrr emits one. ~22 of the
  ~558 vendored defs differ. *arr keys categories by id, so a duplicate id with a
  second name is cosmetic. **`[Accepted]`**
- **Custom-category top-level ordering** — Jackett sorts the top-level tree with
  C# `OrderBy` over the `"zzz"+Name` key, which is **CurrentCulture** (linguistic)
  string comparison; harbrr uses Go byte-**ordinal** `<`. The standard-parent half
  is unaffected (4-digit ids sort identically). For custom categories whose names
  differ by case/punctuation the document order can differ (~309 defs). Only the
  ORDER of top-level `<category>` nodes changes; ids, names, membership and subcats
  are identical, and *arr keys on id, not order. **`[Accepted]`**
- **Same-name custom tie-break** — two customs with identical names tie on the
  sort key; harbrr breaks the tie by ascending id, Jackett by tree-insertion
  order. Sibling order only. **`[Accepted]`**

### Results feed

- **`<jackettindexer>` element name** — kept verbatim for compatibility with
  Torznab consumers that historically scraped Jackett feeds; populated with
  harbrr's indexer id/name. Informational; *arr ignores it. **`[Deliberate]`**
- **`downloadvolumefactor`/`uploadvolumefactor` always emitted** — harbrr's
  normalizer always carries these (defaulting to 1.0), so they are always emitted;
  Jackett omits them when the definition does not extract them. Newznab consumers
  treat an absent factor as 1.0, so an explicit `1` is equivalent. **`[Deliberate]`**
- **`seeders`/`peers` always emitted** — required, non-nullable in harbrr's
  release; Jackett also emits them whenever extracted. **`[Deliberate]`**
- **`files`/`grabs`/`year`/`minimumratio`/`minimumseedtime` omitted at 0** —
  harbrr's non-nullable model cannot distinguish a field that was extracted as 0
  from one that was never present, so 0 is treated as absent and omitted; Jackett
  emits `0` for a present-but-empty non-optional field. A `0` value carries no
  signal a consumer acts on. **`[Accepted]`**
- **Future `pubDate` always clamped to now** — harbrr clamps a future publish date
  to now (`FixResults`); Jackett does this only in release (non-DEBUG) builds.
  harbrr always clamps, matching release-build Jackett. **`[Deliberate]`**
- **`pubDate` timezone** — RFC1123Z preserves the source offset; Jackett renders
  in the host's local offset. Same instant, both valid. **`[Accepted]`**
- **`genre` wire join** — emitted as `", "` (comma+space), matching
  `ResultPage`'s `string.Join(", ", Genres)`; harbrr's internal normalized form
  stays comma-joined (Jackett's filter-facing form). Not a divergence — recorded
  so the two joins are not confused.
- **`language`/`subs` torznab:attrs never emitted** — harbrr's release has no
  language/subs fields, so these attrs are always absent (Jackett omits them when
  null too). **`[Accepted]`**
- **`U+FFFD` handling** — `sanitizeXMLText` strips the Jackett control/BOM/
  noncharacter set and lone surrogates / invalid UTF-8 bytes, but preserves a
  genuine 3-byte `U+FFFD` (which Jackett's regex also preserves). **`[Accepted]`**
- **Download links served direct** — harbrr serves the resolved tracker download/
  magnet link (which legitimately carries a passkey — intended output, never
  logged); it does not yet rewrite links to a proxy `/dl` endpoint or run the
  download resolver. **`[Tracked: Phase 4]`**
