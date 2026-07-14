package mysql

import (
	"reflect"
	"testing"
)

func TestSplitSQLStatements(t *testing.T) {
	script := `
-- regular comment containing ;
CREATE TABLE test (value varchar(32));
INSERT INTO test VALUES ('semi;colon'), ('escaped \' quote');
/* ignored ; comment */
/*!40101 SET character_set_client = utf8mb4 */;
DELIMITER $$
CREATE TRIGGER test_insert AFTER INSERT ON test
FOR EACH ROW BEGIN
  INSERT INTO test VALUES ('trigger;value');
END$$
DELIMITER ;
# another comment
SELECT "done;";
`
	want := []string{
		"CREATE TABLE test (value varchar(32))",
		"INSERT INTO test VALUES ('semi;colon'), ('escaped \\' quote')",
		"/*!40101 SET character_set_client = utf8mb4 */",
		"CREATE TRIGGER test_insert AFTER INSERT ON test\nFOR EACH ROW BEGIN\n  INSERT INTO test VALUES ('trigger;value');\nEND",
		`SELECT "done;"`,
	}
	got, err := splitSQLStatements(script)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("statements:\n%q\nwant:\n%q", got, want)
	}
}

func TestSplitSQLStatementsRejectsUnterminatedInput(t *testing.T) {
	if _, err := splitSQLStatements("SELECT 'unterminated;"); err == nil {
		t.Fatal("unterminated quote was accepted")
	}
}
