// Command seekbrr is a Cardigann-compatible Torznab/Newznab search provider for
// the autobrr family. See docs/ideas.md for the design and docs/plan.md for the
// build order.
package main

import (
	"fmt"
	"os"

	"github.com/autobrr/seekbrr/internal/version"
)

func main() {
	// TODO(seekbrr): wire cobra/viper per AGENTS.md. This is a placeholder
	// entrypoint that keeps `go build ./...` green from the first commit.
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version") {
		fmt.Println(version.String())
		return
	}
	fmt.Fprintf(os.Stderr, "seekbrr %s — not yet implemented; see docs/plan.md\n", version.String())
}
