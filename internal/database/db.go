package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"zion-english/internal/database/queries"
)

type DBMode string

const (
	DB_MODE_RO DBMode = "ro"
	DB_MODE_RW DBMode = "rw"
)

type Service interface {
	Health() map[string]string
	Close() error
	GetQueries() *queries.Queries
	GetDB() *sql.DB
}

type service struct {
	db      *sql.DB
	queries *queries.Queries
}

func (s *service) GetQueries() *queries.Queries {
	return s.queries
}

func (s *service) GetDB() *sql.DB {
	return s.db
}

var (
	dbInstanceRO *service
	dbInstanceRW *service
	dbPath       string
)

func New(mode DBMode) Service {
	switch mode {
	case DB_MODE_RO:
		if dbInstanceRO != nil {
			return dbInstanceRO
		}
	case DB_MODE_RW:
		if dbInstanceRW != nil {
			return dbInstanceRW
		}
	default:
		panic("db mode enum not handled")
	}

	dataSourceName := dbPath + "?_journal_mode=wal" + "&mode=" + string(mode)

	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		fmt.Println("db open error:", err)
		return nil
	}

	switch mode {
	case DB_MODE_RO:
		dbInstanceRO = &service{
			db:      db,
			queries: queries.New(db),
		}
		return dbInstanceRO
	case DB_MODE_RW:
		dbInstanceRW = &service{
			db:      db,
			queries: queries.New(db),
		}
		return dbInstanceRW
	default:
		panic("db mode enum not handled")
	}
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	return stats
}

func (s *service) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

var _ Service = (*service)(nil)

func Init(path string) error {
	dbPath = path

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	goose.SetBaseFS(nil)
	if err := goose.Up(db, "migrations/sqlite3"); err != nil {
		return err
	}

	_ = db.Close()

	_ = New(DB_MODE_RO)
	_ = New(DB_MODE_RW)

	return nil
}

func Close() error {
	if dbInstanceRO != nil {
		if err := dbInstanceRO.Close(); err != nil {
			return err
		}
	}
	if dbInstanceRW != nil {
		if err := dbInstanceRW.Close(); err != nil {
			return err
		}
	}
	return nil
}
