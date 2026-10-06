package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
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
	db   *sql.DB
	conn *sql.Conn
}

func (s *sqlDBSession) Ping(ctx context.Context) error {
	return s.conn.PingContext(ctx)
}

func (s *sqlDBSession) QueryString(ctx context.Context, query string, args ...any) (string, error) {
	var value sql.NullString
	err := s.conn.QueryRowContext(ctx, query, args...).Scan(&value)
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
	_, err := s.conn.ExecContext(ctx, query, args...)
	return err
}

func (s *sqlDBSession) Close() error {
	return errors.Join(s.conn.Close(), s.db.Close())
}

type sessionOpener func(context.Context, *mysql.Config) (dbSession, error)

const (
	connectTimeout    = 10 * time.Second
	reconnectInterval = 5 * time.Second
)

func openDBSession(ctx context.Context, config *mysql.Config) (dbSession, error) {
	connector, err := mysql.NewConnector(config)
	if err != nil {
		return nil, err
	}

	sess, err := openSQLDBSession(ctx, connector)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func openSQLDBSession(ctx context.Context, connector driver.Connector) (*sqlDBSession, error) {
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(3 * time.Minute)

	// SQL scripts may contain START TRANSACTION and session settings. Keep their physical connection reserved until the session is closed:
	// even a one-connection pool can otherwise replace it between statements.
	conn, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &sqlDBSession{db: db, conn: conn}, nil
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

	var sess dbSession
	err = bootstrap.Retry(ctx, bootstrap.RetryOptions{
		Interval: reconnectInterval,
		OnRetry: func(err error) {
			bootstrap.LogDebug(db.env, "**** MySQL connection failed: %v", err)
			bootstrap.LogInfo("**** MySQL server is not available. Waiting %s...", reconnectInterval)
		},
	}, func() error {
		attemptCtx, cancel := context.WithTimeout(ctx, connectTimeout)
		defer cancel()
		opened, err := db.open(attemptCtx, config)
		if err == nil {
			err = opened.Ping(attemptCtx)
		}
		if err != nil {
			if opened != nil {
				_ = opened.Close()
			}

			return err
		}
		sess = opened

		return nil
	})
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (db *DB) Wait() error {
	sess, err := db.waitForConnection(db.user, db.password)
	if err != nil {
		return err
	}
	if err := sess.Close(); err != nil {
		return fmt.Errorf("close MySQL connection: %w", err)
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
