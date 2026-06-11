// Package dbinterface defines the storage interface seekbrr depends on, so a
// backend can be swapped without touching call sites. SQLite is the only
// implementation for now; Postgres is intentionally deferred (see AGENTS.md and
// docs/ideas.md). Do not implement Postgres yet.
package dbinterface
