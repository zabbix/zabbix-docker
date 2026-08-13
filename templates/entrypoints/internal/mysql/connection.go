package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

type dbSession interface {
	Ping(context.Context) error
	QueryString(context.Context, string, ...any) (string, error)
	Exec(context.Context, string, ...any) error
	Close() error
}

type sqlDBSession struct {
	db *sql.DB
}

func (s *sqlDBSession) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *sqlDBSession) QueryString(ctx context.Context, query string, args ...any) (string, error) {
	var value sql.NullString
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if !value.Valid {
		return "", nil
	}

	return value.String, nil
}

func (s *sqlDBSession) Exec(ctx context.Context, query string, args ...any) error {
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *sqlDBSession) Close() error {
	return s.db.Close()
}

type sessionOpener func(*mysql.Config) (dbSession, error)

const connectTimeout = 10 * time.Second

func openDBSession(config *mysql.Config) (dbSession, error) {
	connector, err := mysql.NewConnector(config)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(3 * time.Minute)

	return &sqlDBSession{db: db}, nil
}

func (db *DB) connConfig(dbName, user, password string) (*mysql.Config, error) {
	tlsConfig, err := db.tlsConfig()
	if err != nil {
		return nil, err
	}

	config := mysql.NewConfig()
	config.User = user
	config.Passwd = password
	config.Net = db.network
	config.Addr = db.address
	config.DBName = dbName
	config.Timeout = connectTimeout
	config.InterpolateParams = true
	config.TLS = tlsConfig
	if err := config.Apply(mysql.Charset(db.charset, "")); err != nil {
		return nil, fmt.Errorf("configure database character set: %w", err)
	}

	return config, nil
}

func (db *DB) waitForConnection(user, password string) (dbSession, error) {
	ctx, stop := bootstrap.TerminationContext()
	defer stop()

	return db.waitForConnectionContext(ctx, user, password)
}

func (db *DB) waitForConnectionContext(ctx context.Context, user, password string) (dbSession, error) {
	bootstrap.LogInfo("********************")
	if socket := db.env["DB_SERVER_SOCKET"]; socket != "" {
		bootstrap.LogInfo("* DB_SERVER_SOCKET: %s", socket)
	} else {
		bootstrap.LogInfo("* DB_SERVER_HOST: %s", db.env["DB_SERVER_HOST"])
		bootstrap.LogInfo("* DB_SERVER_PORT: %s", db.env["DB_SERVER_PORT"])
	}
	bootstrap.LogInfo("* DB_SERVER_DBNAME: %s", db.name)
	bootstrap.LogDebug(db.env, "* DB_SERVER_ROOT_USER: %s", db.adminUser)
	bootstrap.LogDebug(db.env, "* DB_SERVER_USER: %s", db.user)
	bootstrap.LogInfo("********************")

	config, err := db.connConfig("", user, password)
	if err != nil {
		return nil, err
	}

	for {
		sess, err := db.open(config)
		if err == nil {
			attemptCtx, cancel := context.WithTimeout(ctx, connectTimeout)
			err = sess.Ping(attemptCtx)
			cancel()
		}
		if err == nil {
			return sess, nil
		}
		if sess != nil {
			_ = sess.Close()
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		bootstrap.LogDebug(db.env, "**** MySQL connection failed: %v", err)

		bootstrap.LogInfo("**** MySQL server is not available. Waiting 5 seconds...")
		timer := time.NewTimer(5 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

// Wait blocks until the database accepts connections with the Zabbix
// credentials.
func (db *DB) Wait() error {
	sess, err := db.waitForConnection(db.user, db.password)
	if err != nil {
		return err
	}
	if err := sess.Close(); err != nil {
		return fmt.Errorf("close database connection: %w", err)
	}

	return nil
}

func (db *DB) query(sess dbSession, query string, args ...any) (string, error) {
	value, err := sess.QueryString(context.Background(), query, args...)
	if err != nil {
		return "", fmt.Errorf("execute database query: %w", err)
	}

	return value, nil
}

func (db *DB) execute(sess dbSession, query string, args ...any) error {
	if err := sess.Exec(context.Background(), query, args...); err != nil {
		return fmt.Errorf("execute database statement: %w", err)
	}

	return nil
}
