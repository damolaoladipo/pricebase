package migrations

import "embed"

// Files holds embedded *.sql migration scripts.
//
//go:embed *.sql
var Files embed.FS
