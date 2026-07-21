package postgresql

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

type dbSession interface {
	QueryString(context.Context, string, ...any) (string, error)
	Exec(context.Context, string, ...any) error
	Close(context.Context) error
}

type pgxDBSession struct {
	conn *pgx.Conn
}

func (s *pgxDBSession) QueryString(ctx context.Context, query string, args ...any) (string, error) {
	var value string
	err := s.conn.QueryRow(ctx, query, args...).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}

	return value, err
}

func (s *pgxDBSession) Exec(ctx context.Context, query string, args ...any) error {
	_, err := s.conn.Exec(ctx, query, args...)
	return err
}

func (s *pgxDBSession) Close(ctx context.Context) error {
	return s.conn.Close(ctx)
}

type connector func(context.Context, *pgx.ConnConfig) (dbSession, error)

func openDBSession(ctx context.Context, config *pgx.ConnConfig) (dbSession, error) {
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return &pgxDBSession{conn: conn}, nil
}

func (db *DB) connConfig(dbName, user, password string) (*pgx.ConnConfig, error) {
	host := db.host
	query := url.Values{}
	if host == "" {
		host = "localhost"
		query.Set("host", "/var/run/postgresql")
	}

	mode := strings.ReplaceAll(db.tls.ConnectMode, "_", "-")
	if mode == "required" {
		mode = "require"
	}
	if mode == "" {
		mode = "disable"
	}
	query.Set("sslmode", mode)

	for _, option := range []struct{ value, parameter string }{
		{db.tls.CAFile, "sslrootcert"},
		{db.tls.CertFile, "sslcert"},
		{db.tls.KeyFile, "sslkey"},
	} {
		if option.value != "" {
			query.Set(option.parameter, option.value)
		}
	}

	connectionURL := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, strconv.Itoa(int(db.port))),
		Path:     "/" + dbName,
		RawQuery: query.Encode(),
	}
	config, err := pgx.ParseConfig(connectionURL.String())
	if err != nil {
		return nil, fmt.Errorf("configure PostgreSQL connection: %w", err)
	}
	config.ConnectTimeout = 10 * time.Second
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	if !db.implicit && db.schema != "" {
		config.RuntimeParams["search_path"] = db.schema
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
	if db.host == "" {
		bootstrap.LogInfo("* DB_SERVER_HOST: Using DB socket")
	} else {
		bootstrap.LogInfo("* DB_SERVER_HOST: %s", db.host)
	}
	bootstrap.LogInfo("* DB_SERVER_PORT: %d", db.port)
	bootstrap.LogInfo("* DB_SERVER_DBNAME: %s", db.name)
	bootstrap.LogInfo("* DB_SERVER_SCHEMA: %s", db.schema)
	bootstrap.LogDebug(db.env, "* DB_SERVER_ZBX_USER: %s", db.user)
	bootstrap.LogInfo("********************")

	for {
		for _, dbName := range []string{user, db.name} {
			config, err := db.connConfig(dbName, user, password)
			if err != nil {
				return nil, err
			}
			attemptCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			sess, err := db.connect(attemptCtx, config)
			cancel()
			if err == nil {
				return sess, nil
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			bootstrap.LogDebug(db.env, "**** PostgreSQL connection to database %q failed: %v", dbName, err)
		}

		bootstrap.LogInfo("**** PostgreSQL server is not available. Waiting 5 seconds...")
		timer := time.NewTimer(5 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func (db *DB) connectTarget(user, password string) (dbSession, error) {
	config, err := db.connConfig(db.name, user, password)
	if err != nil {
		return nil, err
	}

	return db.connect(context.Background(), config)
}

// Wait blocks until the database accepts connections with the Zabbix
// credentials.
func (db *DB) Wait() error {
	sess, err := db.waitForConnection(db.user, db.password)
	if err != nil {
		return err
	}
	if err := sess.Close(context.Background()); err != nil {
		return fmt.Errorf("close PostgreSQL connection: %w", err)
	}

	return nil
}
