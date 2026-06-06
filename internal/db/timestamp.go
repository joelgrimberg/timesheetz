package db

import "time"

// timestampLayout is the canonical format every INSERT/UPDATE writes into
// created_at / updated_at across both SQLite and PostgreSQL. Using one
// Go-supplied string (UTC, second precision, no timezone suffix) avoids the
// SQLite/Postgres CURRENT_TIMESTAMP format mismatch that broke sync's
// lexical timestamp comparison.
const timestampLayout = "2006-01-02 15:04:05"

// NowTimestamp returns the current UTC time formatted for the timestamp
// columns. Tests can monkey-patch nowFunc to control time.
func NowTimestamp() string {
	return nowFunc().UTC().Format(timestampLayout)
}

var nowFunc = time.Now
