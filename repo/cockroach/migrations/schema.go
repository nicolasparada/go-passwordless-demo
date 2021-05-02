package migrations

import (
	_ "embed"
)

//go:embed 000_schema.sql
var Schema string
