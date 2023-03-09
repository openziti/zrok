package store

import (
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
	Path string `cf:"+secret"`
	Type string
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
	if err := store.migrate(cfg); err != nil {
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

func (self *Store) migrate(cfg *Config) error {
	switch cfg.Type {
	case "sqlite3":
		migrations := &migrate.EmbedFileSystemMigrationSource{
			FileSystem: sqlite3_schema.FS,
			Root:       "/",
		}
		migrate.SetTable("migrations")
		n, err := migrate.Exec(self.db.DB, "sqlite3", migrations, migrate.Up)
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
		n, err := migrate.Exec(self.db.DB, "postgres", migrations, migrate.Up)
		if err != nil {
			return errors.Wrap(err, "error running migrations")
		}
		logrus.Infof("applied %d migrations", n)
	}
	return nil
}
