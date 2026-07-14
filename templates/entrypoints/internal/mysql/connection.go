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

type databaseSession interface {
	Ping(context.Context) error
	QueryString(context.Context, string, ...any) (string, error)
	Exec(context.Context, string, ...any) error
	Close() error
}

type sqlDatabaseSession struct {
	db *sql.DB
}

func (s *sqlDatabaseSession) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *sqlDatabaseSession) QueryString(ctx context.Context, query string, args ...any) (string, error) {
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

func (s *sqlDatabaseSession) Exec(ctx context.Context, query string, args ...any) error {
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *sqlDatabaseSession) Close() error {
	return s.db.Close()
}

type sessionOpener func(*mysql.Config) (databaseSession, error)

func openDatabaseSession(config *mysql.Config) (databaseSession, error) {
	connector, err := mysql.NewConnector(config)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(3 * time.Minute)

	return &sqlDatabaseSession{db: db}, nil
}

func (db *Database) connectionConfig(databaseName, user, password string) (*mysql.Config, error) {
	tlsConfig, err := db.tlsConfig()
	if err != nil {
		return nil, err
	}

	config := mysql.NewConfig()
	config.User = user
	config.Passwd = password
	config.Net = db.network
	config.Addr = db.address
	config.DBName = databaseName
	config.Timeout = 10 * time.Second
	config.InterpolateParams = true
	config.TLS = tlsConfig
	if err := config.Apply(mysql.Charset(db.characterSet, "")); err != nil {
		return nil, fmt.Errorf("configure database character set: %w", err)
	}

	return config, nil
}

func (db *Database) waitForConnection(user, password string) (databaseSession, error) {
	bootstrap.LogInfo("********************")
	if socket := db.env["DB_SERVER_SOCKET"]; socket != "" {
		bootstrap.LogInfo("* DB_SERVER_SOCKET: %s", socket)
	} else {
		bootstrap.LogInfo("* DB_SERVER_HOST: %s", db.env["DB_SERVER_HOST"])
		bootstrap.LogInfo("* DB_SERVER_PORT: %s", db.env["DB_SERVER_PORT"])
	}
	bootstrap.LogInfo("* DB_SERVER_DBNAME: %s", db.name)
	bootstrap.LogDebug(db.env, "* DB_SERVER_ROOT_USER: %s", db.rootUser)
	bootstrap.LogDebug(db.env, "* DB_SERVER_ZBX_USER: %s", db.zabbixUser)
	bootstrap.LogInfo("********************")

	config, err := db.connectionConfig("", user, password)
	if err != nil {
		return nil, err
	}

	for {
		sess, err := db.open(config)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err = sess.Ping(ctx)
			cancel()
		}
		if err == nil {
			return sess, nil
		}
		if sess != nil {
			_ = sess.Close()
		}
		bootstrap.LogDebug(db.env, "**** MySQL connection failed: %v", err)

		bootstrap.LogInfo("**** MySQL server is not available. Waiting 5 seconds...")
		db.sleep(5 * time.Second)
	}
}

// Wait blocks until the database accepts connections with the Zabbix
// credentials.
func (db *Database) Wait() error {
	sess, err := db.waitForConnection(db.zabbixUser, db.zabbixPassword)
	if err != nil {
		return err
	}
	if err := sess.Close(); err != nil {
		return fmt.Errorf("close database connection: %w", err)
	}

	return nil
}

func (db *Database) query(sess databaseSession, query string, args ...any) (string, error) {
	value, err := sess.QueryString(context.Background(), query, args...)
	if err != nil {
		return "", fmt.Errorf("execute database query: %w", err)
	}

	return value, nil
}

func (db *Database) execute(sess databaseSession, query string, args ...any) error {
	if err := sess.Exec(context.Background(), query, args...); err != nil {
		return fmt.Errorf("execute database statement: %w", err)
	}

	return nil
}
