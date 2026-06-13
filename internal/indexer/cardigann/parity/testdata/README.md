# Parity fixtures

Each subdirectory here is one parity **case**: a `case.yml` spec plus the files it
references. The harness (`../parity.go`, driven by `../parity_test.go`) runs the
real Cardigann engine over the saved bytes — offline, no network — and
byte-compares the canonical JSON it produces against the case's golden.

## Case layout

```
<case-name>/
  case.yml        # the spec (see fields below)
  definition.yml  # the tracker definition (or use vendor_def to load a vendored one)
  response.html   # a saved response body (parse mode)
  golden.json     # the expected canonical output
```

## `case.yml` fields

- `name` — label (defaults to the directory name)
- `archetype` — the compatibility-matrix row(s) this case covers (required; the
  success-criteria gate asserts every archetype is exercised)
- `golden_source` — provenance of the golden:
  - `jackett-port` — the expected values are Jackett's own test assertions,
    ported verbatim (the authoritative offline oracle)
  - `hand-derived` — values computed by hand from documented Jackett semantics;
    record the derivation reasoning in `description`
- `mode` — `parse` (extract from a saved body; default)
- `definition` / `vendor_def` — set exactly one
- `response` — saved body file (parse mode)
- `response_type` — override the def's response type (`json` / empty)
- `base_url`, `clock` (RFC3339), `config` (the `.Config` namespace), `query`
- `golden` — golden filename (defaults to `golden.json`)

## Oracle policy (offline)

Goldens are **not** captured from a live Jackett (project decision; harbrr is
GPL-2.0, same as Jackett, so porting Jackett's own test material is
license-compatible). They come from Jackett's asserted values (`jackett-port`)
or a written hand-derivation (`hand-derived`). Never blindly `-update` a
`jackett-port` golden — that would let the engine grade its own homework.

## Regenerating goldens

```
go test ./internal/indexer/cardigann/parity/ -run TestParity -update
```

Only after confirming the output matches the case's oracle.
