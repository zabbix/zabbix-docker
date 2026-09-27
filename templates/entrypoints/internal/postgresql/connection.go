package postgresql

import (
	"context"
	"errors"
	"fmt"
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
	Close() error
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

// Close ends the session with its own deadline: the caller is usually done
// with the connection and the terminate message must not block startup.
func (s *pgxDBSession) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), closeTimeout)
	defer cancel()

	return s.conn.Close(ctx)
}

type sessionOpener func(context.Context, *pgx.ConnConfig) (dbSession, error)

// connectTimeoutSeconds is the libpq connect_timeout connection parameter.
const connectTimeoutSeconds = 10

// closeTimeout bounds the terminate message sent when a session is closed.
const closeTimeout = 5 * time.Second

// reconnectInterval is the pause between connection attempts while the
// database is starting up.
const reconnectInterval = 5 * time.Second

func openDBSession(ctx context.Context, config *pgx.ConnConfig) (dbSession, error) {
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return &pgxDBSession{conn: conn}, nil
}

func (db *DB) connConfig(dbName, user, password string) (*pgx.ConnConfig, error) {
	params := url.Values{}
	hosts := make([]string, 0, len(db.endpoints))
	ports := make([]string, 0, len(db.endpoints))
	for _, endpoint := range db.endpoints {
		hosts = append(hosts, endpoint.host)
		ports = append(ports, endpoint.port)
	}
	params.Set("host", strings.Join(hosts, ","))
	params.Set("port", strings.Join(ports, ","))

	params.Set("connect_timeout", strconv.Itoa(connectTimeoutSeconds))
	if len(db.endpoints) > 1 {
		params.Set("target_session_attrs", "read-write")
	}

	mode := strings.ReplaceAll(db.tls.ConnectMode, "_", "-")
	if mode == "required" {
		mode = "require"
	}
	if mode == "" {
		mode = "disable"
	}
	params.Set("sslmode", mode)

	for _, option := range []struct{ value, key string }{
		{db.tls.CAFile, "sslrootcert"},
		{db.tls.CertFile, "sslcert"},
		{db.tls.KeyFile, "sslkey"},
	} {
		if option.value != "" {
			params.Set(option.key, option.value)
		}
	}

	// Host and port query parameters override this placeholder. Keeping the
	// endpoint lists out of URL.Host preserves mixed default and explicit ports.
	connURL := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     "localhost",
		Path:     "/" + dbName,
		RawQuery: params.Encode(),
	}
	config, err := pgx.ParseConfig(connURL.String())
	if err != nil {
		return nil, fmt.Errorf("configure PostgreSQL connection: %w", err)
	}
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	if !db.implicitSearchPath && db.schema != "" {
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
	bootstrap.LogInfo("* DB_SERVER_PORT: %s", db.port)
	bootstrap.LogInfo("* DB_SERVER_DBNAME: %s", db.name)
	bootstrap.LogInfo("* DB_SERVER_SCHEMA: %s", db.schema)
	bootstrap.LogDebug(db.env, "* DB_SERVER_USER: %s", db.user)
	bootstrap.LogInfo("********************")

	var sess dbSession
	err := bootstrap.Retry(ctx, bootstrap.RetryOptions{
		Interval: reconnectInterval,
		OnRetry: func(error) {
			bootstrap.LogInfo("**** PostgreSQL server is not available. Waiting %s...", reconnectInterval)
		},
	}, func() error {
		// The Zabbix user may only be allowed to connect to its own database,
		// so both candidates are tried before the attempt counts as failed.
		var lastErr error
		for _, dbName := range []string{user, db.name} {
			opened, err := db.connect(ctx, dbName, user, password)
			if err == nil {
				sess = opened

				return nil
			}
			lastErr = err
			bootstrap.LogDebug(db.env, "**** PostgreSQL connection to database %q failed: %v", dbName, err)
		}

		return lastErr
	})
	if err != nil {
		return nil, err
	}

	return sess, nil
}

// connect opens one session, bounded by the connect timeout of every
// configured endpoint. A configuration error is not retried.
func (db *DB) connect(ctx context.Context, dbName, user, password string) (dbSession, error) {
	config, err := db.connConfig(dbName, user, password)
	if err != nil {
		return nil, bootstrap.Stop(err)
	}

	timeout := config.ConnectTimeout + time.Second
	if len(db.endpoints) > 1 {
		timeout = time.Duration(len(db.endpoints))*config.ConnectTimeout + time.Second
	}
	attemptCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return db.open(attemptCtx, config)
}

func (db *DB) connectTarget(user, password string) (dbSession, error) {
	return db.connect(context.Background(), db.name, user, password)
}

// Wait blocks until the database accepts connections with the Zabbix
// credentials.
func (db *DB) Wait() error {
	sess, err := db.waitForConnection(db.user, db.password)
	if err != nil {
		return err
	}
	if err := sess.Close(); err != nil {
		return fmt.Errorf("close PostgreSQL connection: %w", err)
	}

	return nil
}
