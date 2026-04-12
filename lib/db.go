package lib

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

// QueryAble is implemented by *sqlx.DB and *sqlx.Tx for repositories.
type QueryAble interface {
	sqlx.Ext
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
}

// Database wraps sqlx.DB.
type Database struct {
	*sqlx.DB
}

// NewDatabase opens PostgreSQL using POSTGRES_* variables.
func NewDatabase(env Env, logger Logger, lc fx.Lifecycle) Database {
	user := url.QueryEscape(env.PostgresUser)
	host := env.PostgresHost
	port := env.PostgresPort
	dbName := env.PostgresDB
	sslMode := env.PostgresSSLMode

	var dsn string
	pass := strings.TrimSpace(env.PostgresPassword)
	if pass != "" {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, url.QueryEscape(pass), host, port, dbName, sslMode,
		)
	} else {
		dsn = fmt.Sprintf(
			"postgres://%s@%s:%s/%s?sslmode=%s",
			user, host, port, dbName, sslMode,
		)
	}

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		logger.Panic(err)
	}
	db.SetMaxOpenConns(20)
	db.SetConnMaxIdleTime(10 * time.Second)
	db.SetMaxIdleConns(2)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		logger.Panicf("database ping failed: %v", err)
	}

	lc.Append(fx.StopHook(func(ctx context.Context) error {
		return db.Close()
	}))

	logger.Info("Database connection established")

	return Database{db}
}
