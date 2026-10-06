package mysql

import (
	"context"
	"database/sql/driver"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

func TestOpenDBSessionConnectionFailure(t *testing.T) {
	const network = "test_open_db_session_connection_failure"
	want := errors.New("database unavailable")
	mysql.RegisterDialContext(network, func(context.Context, string) (net.Conn, error) {
		return nil, want
	})
	t.Cleanup(func() { mysql.DeregisterDialContext(network) })

	config := mysql.NewConfig()
	config.Net = network
	config.Addr = "database:3306"

	sess, err := openDBSession(context.Background(), config)
	if !errors.Is(err, want) {
		t.Fatalf("openDBSession() error = %v, want %v", err, want)
	}
	if sess != nil {
		t.Fatalf("openDBSession() returned a non-nil session on connection failure: %#v", sess)
	}
}

func TestSQLDBSessionKeepsTransactionPastConnectionLifetime(t *testing.T) {
	connector := &sessionTestConnector{}
	ctx := context.Background()
	sess, err := openSQLDBSession(ctx, connector)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sess.Close() })
	sess.db.SetConnMaxLifetime(time.Millisecond)

	if err := sess.Exec(ctx, "START TRANSACTION"); err != nil {
		t.Fatal(err)
	}
	if err := sess.Exec(ctx, "INSERT parent"); err != nil {
		t.Fatal(err)
	}
	// The pool may retire an expired idle connection between statements.
	time.Sleep(10 * time.Millisecond)
	if err := sess.Exec(ctx, "INSERT child"); err != nil {
		t.Fatalf("transaction was lost between statements: %v", err)
	}
	if err := sess.Exec(ctx, "COMMIT"); err != nil {
		t.Fatal(err)
	}
	if connector.opened != 1 {
		t.Fatalf("opened %d connections during one session, want 1", connector.opened)
	}
	if err := sess.Close(); err != nil {
		t.Fatal(err)
	}
	if !connector.last.closed {
		t.Error("session did not close its connection")
	}
}

func TestOpenSQLDBSessionCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	connector := &sessionTestConnector{}
	sess, err := openSQLDBSession(ctx, connector)
	if !errors.Is(err, context.Canceled) || sess != nil {
		t.Fatalf("openSQLDBSession() = %v, %v; want no session and context.Canceled", sess, err)
	}
	if connector.opened != 0 {
		t.Fatalf("opened %d connections with a canceled context", connector.opened)
	}
}

// sessionTestConnector models transaction state belonging to a physical
// connection, not to the database/sql pool that may replace it.
type sessionTestConnector struct {
	opened int
	last   *sessionTestConn
}

func (c *sessionTestConnector) Connect(ctx context.Context) (driver.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.opened++
	c.last = &sessionTestConn{}
	return c.last, nil
}

func (*sessionTestConnector) Driver() driver.Driver { return sessionTestDriver{} }

type sessionTestDriver struct{}

func (sessionTestDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("use the connector")
}

type sessionTestConn struct {
	transaction bool
	parent      bool
	closed      bool
}

func (*sessionTestConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepared statements are not supported")
}

func (*sessionTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are started by the SQL script")
}

func (c *sessionTestConn) Close() error {
	c.closed = true
	c.transaction = false
	c.parent = false
	return nil
}

func (c *sessionTestConn) ExecContext(ctx context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch query {
	case "START TRANSACTION":
		c.transaction = true
	case "INSERT parent":
		if !c.transaction {
			return nil, errors.New("missing transaction")
		}
		c.parent = true
	case "INSERT child":
		if !c.transaction || !c.parent {
			return nil, errors.New("foreign key constraint fails: parent was rolled back")
		}
	case "COMMIT":
		c.transaction = false
	}
	return driver.RowsAffected(1), nil
}
