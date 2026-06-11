package parity

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// update regenerates golden files: `go test ./... -run TestParity -update`.
// Use it only after confirming the new output matches Jackett's.
var update = flag.Bool("update", false, "update parity golden files")

// TestParity runs every saved fixture under testdata/ through the engine and
// diffs the normalized output against the recorded Jackett output. It is the
// gate the engine must pass (docs/ideas.md "Definition of done").
//
// The suite skips cleanly when there are no fixtures, so the baseline build
// stays green; add the first <name>.input + <name>.golden.json pair to switch
// it on. This is the pattern to copy for every stage's fixtures.
func TestParity(t *testing.T) {
	inputs, err := filepath.Glob(filepath.Join("testdata", "*.input"))
	if err != nil {
		t.Fatalf("glob fixtures: %v", err)
	}
	if len(inputs) == 0 {
		t.Skip("no parity fixtures yet — add <name>.input + <name>.golden.json under testdata/ (see AGENTS.md)")
	}

	for _, in := range inputs {
		name := strings.TrimSuffix(filepath.Base(in), ".input")
		t.Run(name, func(t *testing.T) {
			input, err := os.ReadFile(in)
			if err != nil {
				t.Fatalf("read input: %v", err)
			}

			got, err := Process(input)
			if err != nil {
				t.Fatalf("Process(%s): %v", name, err)
			}

			golden := filepath.Join("testdata", name+".golden.json")
			if *update {
				if err := os.WriteFile(golden, got, 0o600); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				return
			}

			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden %s (run with -update once verified): %v", golden, err)
			}
			if string(got) != string(want) {
				t.Errorf("parity mismatch for %s — engine output differs from Jackett's golden (run -update to refresh once verified)", name)
			}
		})
	}
}
