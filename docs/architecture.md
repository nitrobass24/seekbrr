# Architecture

Read this before changing cross-module data flow, service boundaries, API routing, or the engine
pipeline shape. The full plan is in `ideas.md`; this is the load-bearing summary.

## What seekbrr is

The autobrr family's native **Torznab/Newznab + on-demand-search provider** — the slot the family
fills today with Prowlarr. Family-native first, Prowlarr-compatible second.

## The engine is a pipeline

The Cardigann engine (`internal/indexer/cardigann`) is a compiler-style pipeline of small, decoupled,
independently-tested stages:

```
loader → mapper → template → filter → selector → dateparse → regexadapter → login → search → normalizer
                                                                                          → torznab (serialize)
```

Data flows one way. A stage takes typed input and returns typed output; stages do not reach across
each other. Each stage owns its fixtures. This shape is deliberate — it keeps functions small (the
complexity linters enforce it) and makes the parity harness tractable.

## Invariants (do not violate)

1. **Definitions are consumed byte-for-byte.** All behavioral differences live in the engine, never in
   the def files. `vendor/` is read-only (a hook enforces it); overrides go in `dropin/`.
2. **Correctness = parity with Jackett's engine on the same input**, established offline. Per-def-vs-live
   correctness is the corpus's responsibility, not ours.
3. **Two HTTP contracts, kept separate.** Torznab/Newznab (XML, `internal/torznab`) is the *arr-facing
   contract; the OpenAPI surface (`internal/web/swagger`) is seekbrr's own management API. They evolve
   independently.
4. **Regex routing is engine behavior, not a per-def edit.** RE2 by default (ReDoS-safe); regexp2 only
   on the defined triggers (opt-in / non-Latin / compile-fail / .NET-only constructs).
5. **Storage is behind `dbinterface`.** SQLite only for now; Postgres is deliberately deferred — keep
   the interface clean, don't implement it yet.
6. **Secrets never leak.** Redact in all logs/traces; encrypt at rest; never commit.

## Boundaries with the family

- **autobrr** consumes seekbrr's Torznab/Newznab feeds (drop-in for Prowlarr); a future native push is
  the family-only upgrade. seekbrr does not do IRC.
- **qui** manages qBittorrent; seekbrr shares `go-qbittorrent` and pushes grabs. seekbrr does not
  reimplement qBit management.
- **rls** does release-name parsing; adopt it, don't port Prowlarr's Parser.
- **mkbrr/upbrr** own creation/upload; seekbrr only shares the tracker-identity layer.
