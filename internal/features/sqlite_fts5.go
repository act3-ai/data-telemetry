//go:build sqlite_fts5
// +build sqlite_fts5

package features

func init() {
	SqliteFTS5 = true // SqliteFTS5 allows full text search
}
