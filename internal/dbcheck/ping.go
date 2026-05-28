// Package dbcheck provides lightweight PostgreSQL reachability checks
// usable from contexts that cannot import internal/db (such as
// internal/config, where importing db would create an import cycle).
package dbcheck

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// PingPostgresURL opens a transient connection to the given PostgreSQL URL,
// pings it with a 5s timeout, and closes it. It returns the round-trip
// duration on success. The live application connection is not affected.
func PingPostgresURL(url string) (time.Duration, error) {
	start := time.Now()
	d, err := sql.Open("postgres", url)
	if err != nil {
		return 0, err
	}
	defer d.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := d.PingContext(ctx); err != nil {
		return 0, err
	}
	return time.Since(start), nil
}
