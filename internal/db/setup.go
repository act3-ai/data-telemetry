package db

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/runtime"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	"gitlab.com/act3-ai/asce/data/telemetry/internal/features"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha1"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
)

// MigrateDB migrates the DB to the new schema and reprocesses bottles.
func MigrateDB(ctx context.Context, conn *gorm.DB, scheme *runtime.Scheme) error {
	log := logger.FromContext(ctx).WithGroup("migrator")

	con := conn.WithContext(logger.NewContext(ctx, log))
	err := con.AutoMigrate(
		&Data{},
		&Digest{},
		&Bottle{},
		&Event{},
		&Blob{},
		&PublicArtifact{},
		&Source{},
		&Deprecates{},
		&Part{},
		&Author{},
		&Label{},
		&Annotation{},
		&Metric{},
		&Manifest{},
		&Layer{},
		&Signature{},
		&SignatureAnnotation{},
	)
	if err != nil {
		return fmt.Errorf("database migration: %w", err)
	}

	switch con.Name() {
	case "postgres":
		err := setupPostgresSearch(log, conn)
		if err != nil {
			return fmt.Errorf("could not run postgres search setup: %w", err)
		}
	case "sqlite":
		if features.SqliteFTS5 {
			err := setupSqliteFTS(log, conn)
			if err != nil {
				return fmt.Errorf("could not run sqlite full text search setup: %w", err)
			}
		}
	}

	// TODO this should only be called by the leader (if we have multiple apps running against one DB)
	processors := []Processor{
		&BlobProcessor{},
		NewBottleProcessor(scheme),
		&ManifestProcessor{},
		&EventProcessor{},
		&SignatureProcessor{},
	}
	for _, processor := range processors {
		if err := Reprocess(ctx, con, processor); err != nil {
			return err
		}
	}

	return nil
}

func setupPostgresSearch(log *slog.Logger, conn *gorm.DB) error {
	ctx := conn.Statement.Context
	log.InfoContext(ctx, "Setting up postgres search")

	if !conn.Migrator().HasColumn(&Bottle{}, "description_tsv") {
		result := conn.Exec(`ALTER TABLE bottles ADD COLUMN description_tsv tsvector GENERATED ALWAYS AS (to_tsvector('english', coalesce(description,''))) STORED;`)
		if result.Error != nil {
			return result.Error
		}
	}

	// Use the GIN index (it is faster)
	// https://stackoverflow.com/questions/12933805/best-way-to-use-postgresql-full-text-search-ranking
	if !conn.Migrator().HasIndex(&Bottle{}, "idx_bottles_description_tsv") {
		result := conn.Exec(`CREATE INDEX idx_bottles_description_tsv ON bottles USING gin(description_tsv);`)
		if result.Error != nil {
			return result.Error
		}
	}

	// conn.Exec(`ALTER TABLE authors ADD COLUMN ts tsvector GENERATED ALWAYS AS (setweight(to_tsvector('english', coalesce(name,'')), 'A') || setweight(to_tsvector('english', coalesce(email,'')), 'B')) STORED;`)
	// conn.Exec(`ALTER TABLE labels ADD COLUMN ts tsvector GENERATED ALWAYS AS (setweight(to_tsvector('english', coalesce(key,'')), 'A') || setweight(to_tsvector('english', coalesce(value,'')), 'B')) STORED;`)
	// conn.Exec(`CREATE VIEW search AS select data, digest, ( setweight(bottles.ts, 'A') || setweight(labels.ts, 'B') || setweight(authors.ts, 'C') ) AS ts FROM bottles FULL OUTER JOIN labels ON bottles.id = labels.bottle_id FULL OUTER JOIN sources ON bottles.id = sources.bottle_id FULL OUTER JOIN authors ON bottles.id = authors.bottle_id;`)
	return nil
}

func setupSqliteFTS(log *slog.Logger, conn *gorm.DB) error {
	ctx := conn.Statement.Context
	log.InfoContext(ctx, "Setting up sqlite full text search")

	if !conn.Migrator().HasTable("description_fts") {
		result := conn.Exec(`CREATE VIRTUAL TABLE description_fts USING fts5(description, content='bottles', content_rowid='id');`)
		if result.Error != nil {
			return result.Error
		}
		result = conn.Exec(`CREATE TRIGGER bottles_fts_ai AFTER INSERT ON bottles BEGIN INSERT INTO description_fts(rowid, description) VALUES(new.id, new.description); END;`)
		if result.Error != nil {
			return result.Error
		}
		result = conn.Exec(
			`CREATE TRIGGER bottles_fts_ad AFTER DELETE ON bottles BEGIN INSERT INTO description_fts(description_fts, rowid, description) VALUES('delete', old.id, old.description); END;`,
		)
		if result.Error != nil {
			return result.Error
		}
		result = conn.Exec(
			`CREATE TRIGGER bottles_fts_au AFTER UPDATE ON bottles BEGIN INSERT INTO description_fts(description_fts, rowid, description) VALUES('delete', old.id, old.description); INSERT INTO description_fts(rowid) VALUES(new.id); END;`,
		)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

// connStr is something like "host=localhost user=postgres password=password dbname=test port=5432 sslmode=disable"

// OpenPostgresDB connect to a Postgres database.
func OpenPostgresDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: NewGormLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("opening Postgres database: %w", err)
	}

	return db, nil
}

// OpenSqliteDB connect to a sqlite database (file).
func OpenSqliteDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: NewGormLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("opening SQLite database: %w", err)
	}

	return db, nil
}

// Open opens a DB connection using URL formatting.
// Also migrates the database
// For example file:test.db, "file::memory:" or postgres://jack:secret@foo.example.com:5432,bar.example.com:5432/mydb
func Open(ctx context.Context, conf v1alpha2.Database, scheme *runtime.Scheme) (*gorm.DB, error) {
	log := logger.FromContext(ctx)

	u, e := url.Parse(string(conf.DSN))
	if e != nil {
		return nil, fmt.Errorf("parsing DSN: %w", e)
	}

	if conf.Password != "" {
		u.User = url.UserPassword(u.User.Username(), string(conf.Password))
	}
	log.InfoContext(ctx, "Database connection", "dsn", u.Redacted())

	var db *gorm.DB
	var err error
	switch u.Scheme {
	case "postgres":
		db, err = OpenPostgresDB(u.String())
	case "file":
		db, err = OpenSqliteDB(u.String())
	default:
		return nil, fmt.Errorf("unknown database type \"%s\"", u.Scheme)
	}
	if err != nil {
		return nil, err
	}

	db = db.WithContext(ctx)

	if err := MigrateDB(ctx, db, scheme); err != nil {
		return nil, err
	}

	return db, nil
}

// Open opens a DB connection using URL formatting.
// Also migrates the database
// For example file:test.db, "file::memory:" or postgres://jack:secret@foo.example.com:5432,bar.example.com:5432/mydb
// uses v1alpha1 db config
func Openv1alpha1(ctx context.Context, conf v1alpha1.Database, scheme *runtime.Scheme) (*gorm.DB, error) {
	log := logger.FromContext(ctx)

	u, e := url.Parse(string(conf.DSN))
	if e != nil {
		return nil, fmt.Errorf("parsing DSN: %w", e)
	}

	if conf.Password != "" {
		u.User = url.UserPassword(u.User.Username(), string(conf.Password))
	}
	log.InfoContext(ctx, "Database connection", "dsn", u.Redacted())

	var db *gorm.DB
	var err error
	switch u.Scheme {
	case "postgres":
		db, err = OpenPostgresDB(u.String())
	case "file":
		db, err = OpenSqliteDB(u.String())
	default:
		return nil, fmt.Errorf("unknown database type \"%s\"", u.Scheme)
	}
	if err != nil {
		return nil, err
	}

	db = db.WithContext(ctx)

	if err := MigrateDB(ctx, db, scheme); err != nil {
		return nil, err
	}

	return db, nil
}
