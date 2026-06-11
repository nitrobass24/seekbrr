// Package parity holds the differential/golden test harness that pins seekbrr's
// engine output to Jackett's. Fixtures live under testdata/: a pair of
// <name>.input (a saved tracker response) and <name>.golden.json (Jackett's
// normalized output for the same input). This suite is the gate the engine must
// pass — see docs/ideas.md "Definition of done" and AGENTS.md.
package parity

import "errors"

// ErrNotImplemented is returned by Process until the cardigann pipeline is
// wired. The parity suite skips cleanly while there are no fixtures, and fails
// loudly once fixtures exist — which is the point: it drives the engine.
var ErrNotImplemented = errors.New("cardigann engine not yet implemented")

// Process runs the engine over a saved tracker response and returns the
// normalized result as canonical JSON, byte-comparable against a golden file.
//
// TODO(seekbrr): invoke the cardigann pipeline (loader -> mapper -> ... ->
// normalizer) and marshal canonically. Keep output deterministic (stable field
// order, normalized whitespace) so golden comparison is meaningful.
func Process(input []byte) ([]byte, error) {
	_ = input
	return nil, ErrNotImplemented
}
