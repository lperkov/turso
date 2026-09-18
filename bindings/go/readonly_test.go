package turso

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestReadOnlyDSN(t *testing.T) {
	for _, value := range []string{"true", "1", "false", "0"} {
		config, err := parseDSN("test.db?readonly=" + value)
		if err != nil {
			t.Fatal(err)
		}
		if config.ReadOnly != (value == "true" || value == "1") {
			t.Fatalf("readonly=%s parsed as %v", value, config.ReadOnly)
		}
	}
	if _, err := parseDSN("test.db?readonly=maybe"); err == nil {
		t.Fatal("invalid readonly accepted")
	}
}

func TestReadOnlyRejectsWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "read.db")
	writer, err := sql.Open("turso", path+"?experimental=multiprocess_wal")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Exec("CREATE TABLE items (value TEXT)"); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Exec("INSERT INTO items VALUES ('original')"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := sql.Open("turso", path+"?experimental=multiprocess_wal&readonly=true")
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var value string
	if err := reader.QueryRow("SELECT value FROM items").Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != "original" {
		t.Fatalf("value=%q", value)
	}
	if _, err := reader.Exec("INSERT INTO items VALUES ('forbidden')"); !errors.Is(err, ErrTursoReadOnly) {
		t.Fatalf("write error=%v", err)
	}
}
