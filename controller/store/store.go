package store

import (
	"fmt"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	postgresql_schema "github.com/openziti/zrok/controller/store/sql/postgresql"
	sqlite3_schema "github.com/openziti/zrok/controller/store/sql/sqlite3"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

type Model struct {
	Id        int
	CreatedAt time.Time
	UpdatedAt time.Time
	Deleted   bool
}

type Config struct {
	Path                 string `df:",secret"`
	Type                 string
	EnableLocking        bool
	DisableAutoMigration bool
}

type Store struct {
	cfg *Config
	db  *sqlx.DB
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
	return store, nil
}

func (str *Store) Begin() (*sqlx.Tx, error) {
	return str.db.Beginx()
}

func (str *Store) Close() error {
	return str.db.Close()
}

func (str *Store) migrate(cfg *Config) error {
	return str.migrateWithDirection(cfg, migrate.Up, 0)
}

func (str *Store) MigrateDown(cfg *Config, max int) error {
	return str.migrateWithDirection(cfg, migrate.Down, max)
}

func (str *Store) migrateWithDirection(cfg *Config, direction migrate.MigrationDirection, max int) error {
	switch cfg.Type {
	case "sqlite3":
		migrations := &migrate.EmbedFileSystemMigrationSource{
			FileSystem: sqlite3_schema.FS,
			Root:       "/",
		}
		migrate.SetTable("migrations")
		n, err := migrate.ExecMax(str.db.DB, "sqlite3", migrations, direction, max)
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
		n, err := migrate.ExecMax(str.db.DB, "postgres", migrations, direction, max)
		if err != nil {
			return errors.Wrap(err, "error running migrations")
		}
		logrus.Infof("applied %d migrations", n)
	}
	return nil
}
