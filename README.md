# seekbrr

A lightweight, Go-native, **Cardigann-compatible Torznab/Newznab search provider** for the autobrr
ecosystem — the family's native indexer search/scrape provider (the slot currently filled by
Prowlarr), built on Cardigann compatibility instead of a new tracker format.

> Status: **pre-build scaffold.** The engine is built test-harness-first; the central bet is
> behavioral parity with Jackett's Cardigann engine on saved inputs. See `docs/ideas.md` for the full
> plan and `docs/plan.md` for the build order. **Read `AGENTS.md` before contributing.**

## Quick start (development)

```sh
make tools         # install gofumpt, goimports, golangci-lint
make vendor-defs   # fetch the Jackett definition snapshot into internal/indexer/definitions/vendor
make build         # -> bin/seekbrr
make test          # go test -race -count=1 ./...
make precommit     # fmt + lint + test (run before final on any change)
pre-commit install # enable gitleaks + lint + secret guard on commit
```

The baseline scaffold builds and tests green with no external dependencies; deps are added as engine
stages land.

## Layout

- `cmd/seekbrr` — entrypoint
- `internal/indexer/cardigann` — the engine, a compiler-style pipeline (one package per stage:
  `loader → … → normalizer`); parity gate in `parity/`
- `internal/indexer/definitions` — embedded Jackett snapshot (`vendor/`, **read-only**) + `dropin/`
  overrides
- `internal/torznab` — the Torznab/Newznab serializer (the *arr-facing contract)
- `docs/ideas.md` — full design & technical plan · `docs/plan.md` — build checklist
- `AGENTS.md` — rules for AI agents and contributors (`CLAUDE.md` is a symlink to it)

## Module path

The Go module is `github.com/autobrr/seekbrr` (the intended home) even while the repo lives at
`nitrobass24/seekbrr` during early development — this avoids an import-path rename on the eventual
move to the autobrr org. Local builds work regardless of where the repo is hosted.

## License

GPL-2.0-or-later. Tracker definitions are vendored from Jackett (GPL-2.0); see `docs/ideas.md`
"Licensing".
