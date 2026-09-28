package store

import (
	"context"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	postgresql_schema "github.com/openziti/zrok/controller/store/sql/postgresql"
	sqlite3_schema "github.com/openziti/zrok/controller/store/sql/sqlite3"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"time"
)

type Model struct {
	Id        int
	CreatedAt time.Time
	UpdatedAt time.Time
	Deleted   bool
}

type Config struct {
	Path                 string `cf:"+secret"`
	Type                 string
	EnableLocking        bool
	DisableAutoMigration bool
}

type Store struct {
	cfg        *Config
	db         *sqlx.DB
	v2Mappings bool
}

func Open(cfg *Config) (*Store, error) {
	var dbx *sqlx.DB
	var err error

	switch cfg.Type {
	case "sqlite3":
		dbx, err = sqlx.Open("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", cfg.Path))
		if err != nil {
			return nil, errors.Wrapf(err, "error opening database '%v'", cfg.Path)
		}
		dbx.DB.SetMaxOpenConns(1)

	case "postgres":
		dbx, err = sqlx.Connect("postgres", cfg.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "error opening database '%v'", cfg.Path)
		}

	default:
		return nil, errors.Errorf("unknown database type '%v' (supported: sqlite3, postgres)", cfg.Type)
	}
	logrus.Info("database connected")
	dbx.MapperFunc(strcase.ToSnake)

	store := &Store{cfg: cfg, db: dbx}
	if !cfg.DisableAutoMigration {
		if err := store.migrate(cfg); err != nil {
			return nil, errors.Wrapf(err, "error migrating database '%v'", cfg.Path)
		}
	}
	store.probeV2Mappings()
	return store, nil
}

// v2MappingTables are the v2 line's name and frontend mapping tables; a store shared with a v2 controller has them,
// a store only ever migrated by v1 does not, and v1 never creates them.
var v2MappingTables = []string{"names", "share_name_mappings", "frontend_mappings"}

// probeV2Mappings records whether all of the v2 mapping tables exist, by attempting a trivial read from each; the
// probe is the same statement on every engine. each read runs outside any transaction, so a failure cannot abort one.
func (str *Store) probeV2Mappings() {
	str.v2Mappings = true
	for _, table := range v2MappingTables {
		rows, err := str.db.Query(fmt.Sprintf("select 1 from %s limit 1", table))
		if err != nil {
			logrus.Infof("v2 mapping table '%v' not present; share teardown leaves v2 mappings alone: %v", table, err)
			str.v2Mappings = false
			return
		}
		_ = rows.Close()
	}
	logrus.Info("v2 mapping tables present; share teardown releases v2 names and mappings")
}

func (str *Store) Begin() (*sqlx.Tx, error) {
	return str.BeginContext(context.Background())
}

func (str *Store) BeginContext(ctx context.Context) (*sqlx.Tx, error) {
	return str.db.BeginTxx(ctx, nil)
}

func (str *Store) Close() error {
	return str.db.Close()
}

func (str *Store) migrate(cfg *Config) error {
	switch cfg.Type {
	case "sqlite3":
		migrations := &migrate.EmbedFileSystemMigrationSource{
			FileSystem: sqlite3_schema.FS,
			Root:       "/",
		}
		migrate.SetTable("migrations")
		n, err := migrate.Exec(str.db.DB, "sqlite3", migrations, migrate.Up)
		if err != nil {
			return errors.Wrap(err, "error running migrations")
		}
		logrus.Infof("applied %d migrations", n)

	case "postgres":
		migrations := &migrate.EmbedFileSystemMigrationSource{
			FileSystem: postgresql_schema.FS,
			Root:       "/",
		}
		migrate.SetTable("migrations")
		n, err := migrate.Exec(str.db.DB, "postgres", migrations, migrate.Up)
		if err != nil {
			return errors.Wrap(err, "error running migrations")
		}
		logrus.Infof("applied %d migrations", n)
	}
	return nil
}
