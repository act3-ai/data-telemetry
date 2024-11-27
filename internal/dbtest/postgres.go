// Package dbtest contains helper functions to interact with the DB for testing purposes only.
package dbtest

import (
	"fmt"
	"net/url"

	"gitlab.com/act3-ai/asce/data/telemetry/v2/internal/db"
)

// CreateTempPostgresDb creates a new database in a postgres instance with a given name and dsn and returns the new dsn and a cleanup function.
func CreateTempPostgresDb(dbName string, dsn string) (string, func(), error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", nil, fmt.Errorf("parsing DSN for temporary database: %w", err)
	}
	testDbName := u.Path[1:] + dbName

	// Connect to default postgres database to create new test db
	u.Path = "postgres"
	pgdb, err := db.OpenPostgresDB(u.String())
	if err != nil {
		return "", nil, err
	}
	u.Path = testDbName

	// NOTE: SQL identifiers that are case-sensitive need to be enclosed in double quotes
	pgdb.Exec("DROP DATABASE IF EXISTS \"" + testDbName + "\"")
	pgdb.Exec("CREATE DATABASE \"" + testDbName + "\"")

	cleanup := func() {
		// drop any connections to the testdb
		// NOTE: assuming postgres version > 9.2
		testDb, err := db.OpenPostgresDB(u.String())
		if err != nil {
			// TODO: surface this error in a better way
			panic(err.Error())
		}
		testDb.Exec("SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE datname = '" + testDbName + "' AND pid <> pg_backend_pid();")
		d, err := testDb.DB()
		if err != nil {
			// TODO: surface this error in a better way
			panic(err.Error())
		}
		d.Close()

		pgdb.Exec("DROP DATABASE IF EXISTS \"" + testDbName + "\"")
	}

	return u.String(), cleanup, nil
}
