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

type sessionOpener func(context.Context, *pgx.ConnConfig) (dbSession, error)

const connectTimeout = 10

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

	params.Set("connect_timeout", strconv.Itoa(connectTimeout))
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

	for {
		for _, dbName := range []string{user, db.name} {
			config, err := db.connConfig(dbName, user, password)
			if err != nil {
				return nil, err
			}
			timeout := config.ConnectTimeout + time.Second
			if len(db.endpoints) > 1 {
				timeout = time.Duration(len(db.endpoints))*config.ConnectTimeout + time.Second
			}
			attemptCtx, cancel := context.WithTimeout(ctx, timeout)
			sess, err := db.open(attemptCtx, config)
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

	return db.open(context.Background(), config)
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
