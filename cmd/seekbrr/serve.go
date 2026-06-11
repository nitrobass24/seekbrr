package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/autobrr/seekbrr/internal/config"
	"github.com/autobrr/seekbrr/internal/logger"
)

// newServeCmd runs the seekbrr server. The HTTP surface, database, and engine
// are wired in later phases (see docs/plan.md); for now serve proves the
// config/logging path: it loads and validates configuration, warns loudly when
// secrets would be stored as plaintext, logs the redacted config, and exits.
func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run the seekbrr server",
		Args:  cobra.NoArgs,
		RunE:  runServe,
	}
}

func runServe(cmd *cobra.Command, _ []string) error {
	cfgFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("read --config flag: %w", err)
	}

	cfg, err := config.Load(cfgFile, cmd.Flags())
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log, err := logger.New(cfg.Log, cmd.OutOrStdout())
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	if !cfg.HasSecretKey() {
		log.Warn().Msg("secrets: no encryption key configured — secrets will be stored in PLAINTEXT (set secrets.encryption_key or secrets.key_file)")
	}
	log.Info().Stringer("config", cfg).Msg("seekbrr starting")
	log.Info().Msg("server not yet implemented (see docs/plan.md)")
	return nil
}
