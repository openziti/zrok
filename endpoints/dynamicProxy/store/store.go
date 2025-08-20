package store

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/endpoints/dynamicProxy/store/sql/postgres"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

type Model struct {
	Id        int
	CreatedAt time.Time
}

type Config struct {
	Path                 string `df:",secret"`
	DisableAutoMigration bool
}

type Store struct {
	cfg *Config
	db  *sqlx.DB
}

func Open(cfg *Config) (*Store, error) {
	dbx, err := sqlx.Connect("postgres", cfg.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening database '%v'", cfg.Path)
	}
	logrus.Info("database connected")

	store := &Store{cfg: cfg, db: dbx}
	if !cfg.DisableAutoMigration {
		if err := store.migrate(); err != nil {
			return nil, errors.Wrapf(err, "error migrating database '%v'", cfg.Path)
		}
	}

	return store, nil
}

func (str *Store) migrate() error {
	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: postgres.FS,
		Root:       "/",
	}
	migrate.SetTable("migrations")
	n, err := migrate.Exec(str.db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		return errors.Wrapf(err, "error applying migrations")
	}
	logrus.Infof("applied %d migrations", n)
	return nil
}
