package store

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/openziti-test-kitchen/zrok/controller/store/sql"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"time"
)

type Model struct {
	Id        int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Config struct {
	Path string
}

type Store struct {
	cfg *Config
	db  *sqlx.DB
}

func Open(cfg *Config) (*Store, error) {
	dbx, err := sqlx.Open("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", cfg.Path))
	if err != nil {
		return nil, errors.Wrapf(err, "error opening database '%v'", cfg.Path)
	}
	dbx.DB.SetMaxOpenConns(1)
	logrus.Infof("opened database '%v'", cfg.Path)
	dbx.MapperFunc(strcase.ToSnake)
	store := &Store{cfg: cfg, db: dbx}
	if err := store.migrate(); err != nil {
		return nil, errors.Wrapf(err, "error migrating database '%v'", cfg.Path)
	}
	return store, nil
}

func (self *Store) Begin() (*sqlx.Tx, error) {
	return self.db.Beginx()
}

func (self *Store) Close() error {
	return self.db.Close()
}

func (self *Store) migrate() error {
	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: sql.Fs,
		Root:       "/",
	}
	migrate.SetTable("migrations")
	n, err := migrate.Exec(self.db.DB, "sqlite3", migrations, migrate.Up)
	if err != nil {
		return errors.Wrap(err, "error running migrations")
	}
	logrus.Infof("applied %d migrations", n)
	return nil
}
