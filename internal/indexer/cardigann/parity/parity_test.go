package parity

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// update regenerates golden files: `go test ./... -run TestParity -update`.
// Use it only after confirming the new output matches the case's oracle
// (Jackett's asserted values for jackett-port cases, the hand-derivation in the
// case description for hand-derived cases). Never refresh a jackett-port golden
// blindly — that would let the engine grade its own homework.
var update = flag.Bool("update", false, "update parity golden files")

// TestParity runs every case directory under testdata/ through the real engine
// and diffs the canonical JSON against the recorded golden. It is the gate the
// engine must pass (docs/ideas.md "Definition of done").
//
// The suite skips cleanly when there are no cases, so the baseline build stays
// green; add the first testdata/<name>/case.yml to switch it on.
func TestParity(t *testing.T) {
	t.Parallel()

	dirs, err := caseDirs()
	if err != nil {
		t.Fatalf("scanning cases: %v", err)
	}
	if len(dirs) == 0 {
		t.Skip("no parity cases yet — add testdata/<name>/case.yml (see testdata/README.md)")
	}

	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()
			runCase(t, dir)
		})
	}
}

// runCase loads, runs, and golden-compares one case directory.
func runCase(t *testing.T, dir string) {
	t.Helper()

	c, err := Load(dir)
	if err != nil {
		t.Fatalf("loading case: %v", err)
	}

	got, err := c.Run(dir)
	if err != nil {
		t.Fatalf("running case %q: %v", c.Name, err)
	}

	golden := filepath.Join(dir, c.GoldenFile())
	if *update {
		if err := os.WriteFile(golden, got, 0o600); err != nil {
			t.Fatalf("writing golden: %v", err)
		}
		return
	}

	want, err := os.ReadFile(golden) //nolint:gosec // golden path under testdata/.
	if err != nil {
		t.Fatalf("reading golden %s (run with -update once the output is verified against the oracle): %v", golden, err)
	}
	if string(got) != string(want) {
		t.Errorf("parity mismatch for %q [archetype=%s source=%s]\n--- got ---\n%s\n--- want ---\n%s",
			c.Name, c.Archetype, c.GoldenSource, got, want)
	}
}

// caseDirs returns every testdata subdirectory holding a case.yml.
func caseDirs() ([]string, error) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var dirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join("testdata", e.Name())
		if _, err := os.Stat(filepath.Join(dir, "case.yml")); err == nil {
			dirs = append(dirs, dir)
		}
	}
	return dirs, nil
}
