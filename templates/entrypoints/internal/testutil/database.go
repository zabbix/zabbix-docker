package testutil

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// FakeDBSession records SQL statements and returns configured query results.
type FakeDBSession struct {
	Queries    map[string]string
	Statements []string
	Closed     bool
}

func (s *FakeDBSession) Ping(ctx context.Context) error { return ctx.Err() }

func (s *FakeDBSession) QueryString(_ context.Context, query string, args ...any) (string, error) {
	s.Statements = append(s.Statements, formatStatement(query, args...))
	return s.Queries[query], nil
}

func (s *FakeDBSession) Exec(_ context.Context, query string, args ...any) error {
	s.Statements = append(s.Statements, formatStatement(query, args...))
	return nil
}

func (s *FakeDBSession) Close() error {
	s.Closed = true
	return nil
}

func formatStatement(query string, args ...any) string {
	if len(args) == 0 {
		return query
	}
	return fmt.Sprintf("%s %v", query, args)
}

// ImportedStatements joins the SQL recorded by import sessions.
func ImportedStatements(sessions []*FakeDBSession) string {
	var statements []string
	for _, s := range sessions {
		statements = append(statements, s.Statements...)
	}
	return strings.Join(statements, "\n")
}

// CheckStatements fails the test unless statements contain every expected statement.
func CheckStatements(t *testing.T, kind, statements string, expected ...string) {
	t.Helper()
	for _, statement := range expected {
		if !strings.Contains(statements, statement) {
			t.Fatalf("%s statements do not contain %q:\n%s", kind, statement, statements)
		}
	}
}
