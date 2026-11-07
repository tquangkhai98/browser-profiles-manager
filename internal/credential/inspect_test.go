package credential

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestInspect_EmptyProfile(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := Inspect(tmpDir, "empty-profile")
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}

	if result == nil {
		t.Fatal("Inspect() returned nil result")
	}
	if result.ProfileName != "empty-profile" {
		t.Errorf("ProfileName = %q, want %q", result.ProfileName, "empty-profile")
	}
	if len(result.Sites) != 0 {
		t.Errorf("expected 0 sites, got %d", len(result.Sites))
	}
	if result.TotalCookies != 0 {
		t.Errorf("TotalCookies = %d, want 0", result.TotalCookies)
	}
	if result.TotalLogins != 0 {
		t.Errorf("TotalLogins = %d, want 0", result.TotalLogins)
	}
}

func TestInspect_NoDatabases(t *testing.T) {
	tmpDir := t.TempDir()
	// Create Default subdirectory but no DB files
	defaultDir := filepath.Join(tmpDir, "Default")
	if err := mkdirAll(defaultDir); err != nil {
		t.Fatalf("cannot create Default dir: %v", err)
	}

	result, err := Inspect(tmpDir, "no-db-profile")
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if len(result.Sites) != 0 {
		t.Errorf("expected 0 sites, got %d", len(result.Sites))
	}
}

func TestFindDBPath_PreferDefault(t *testing.T) {
	tmpDir := t.TempDir()

	// Create both locations
	defaultDir := filepath.Join(tmpDir, "Default")
	if err := mkdirAll(defaultDir); err != nil {
		t.Fatal(err)
	}

	// File only in Default/
	defaultCookies := filepath.Join(defaultDir, "Cookies")
	if err := writeTestFile(defaultCookies, []byte("default-cookies")); err != nil {
		t.Fatal(err)
	}

	found := findDBPath(tmpDir, "Cookies")
	if found != defaultCookies {
		t.Errorf("findDBPath() = %q, want %q", found, defaultCookies)
	}
}

func TestFindDBPath_FallbackToRoot(t *testing.T) {
	tmpDir := t.TempDir()

	// File only at root
	rootCookies := filepath.Join(tmpDir, "Cookies")
	if err := writeTestFile(rootCookies, []byte("root-cookies")); err != nil {
		t.Fatal(err)
	}

	found := findDBPath(tmpDir, "Cookies")
	if found != rootCookies {
		t.Errorf("findDBPath() = %q, want %q", found, rootCookies)
	}
}

func TestFindDBPath_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	found := findDBPath(tmpDir, "NonExistent")
	if found != "" {
		t.Errorf("findDBPath() = %q, want empty string", found)
	}
}

func TestFindAllDBPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files at both locations
	defaultDir := filepath.Join(tmpDir, "Default")
	if err := mkdirAll(defaultDir); err != nil {
		t.Fatal(err)
	}

	defaultPath := filepath.Join(defaultDir, "Cookies")
	rootPath := filepath.Join(tmpDir, "Cookies")
	if err := writeTestFile(defaultPath, []byte("default")); err != nil {
		t.Fatal(err)
	}
	if err := writeTestFile(rootPath, []byte("root")); err != nil {
		t.Fatal(err)
	}

	found := findAllDBPaths(tmpDir, "Cookies")
	if len(found) != 2 {
		t.Errorf("findAllDBPaths() returned %d paths, want 2", len(found))
	}
}

// Helper functions
func mkdirAll(path string) error {
	return os.MkdirAll(path, 0700)
}

func writeTestFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0600)
}

func TestInspect_ReadOnlyAndCounts(t *testing.T) {
	tmpDir := t.TempDir()
	if err := mkdirAll(filepath.Join(tmpDir, "Default")); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmpDir, "Default", "Cookies")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE cookies (host_key TEXT); INSERT INTO cookies VALUES ('a.com'), ('a.com'), ('b.com')`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if err := os.Chmod(dbPath, 0400); err != nil {
		t.Fatal(err)
	}

	result, err := Inspect(tmpDir, "p")
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.TotalCookies != 3 {
		t.Errorf("TotalCookies = %d, want 3", result.TotalCookies)
	}
	// A read-only open must not create a journal next to the database.
	if _, err := os.Stat(dbPath + "-journal"); !os.IsNotExist(err) {
		t.Errorf("journal created by Inspect: %v", err)
	}
}

func TestInspect_AllDatabasesFail(t *testing.T) {
	tmpDir := t.TempDir()
	if err := writeTestFile(filepath.Join(tmpDir, "Cookies"), []byte("not a sqlite database")); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(tmpDir, "broken"); err == nil {
		t.Fatal("expected error when every credential DB fails to open")
	}
}

func TestReadOnlyDSN(t *testing.T) {
	got := readOnlyDSN("/tmp/a b/Login Data")
	want := "file:/tmp/a%20b/Login%20Data?mode=ro&immutable=1"
	if got != want {
		t.Errorf("readOnlyDSN = %q, want %q", got, want)
	}
}
