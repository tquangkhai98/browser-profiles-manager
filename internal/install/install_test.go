package install

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteConfigBacksUpAndKeepsSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks need privileges on Windows")
	}
	dir := t.TempDir()
	real := filepath.Join(dir, "dotfiles", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(real), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(real, []byte(`{"keep":1}`), 0640); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "mcp.json")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}

	if err := writeConfig(link, map[string]interface{}{"mcpServers": map[string]interface{}{}}); err != nil {
		t.Fatalf("writeConfig: %v", err)
	}

	if fi, err := os.Lstat(link); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("symlink replaced: %v %v", fi, err)
	}
	if fi, err := os.Stat(real); err != nil || fi.Mode().Perm() != 0640 {
		t.Fatalf("mode not preserved: %v %v", fi, err)
	}
	bak, err := os.ReadFile(real + ".bpm.bak")
	if err != nil || string(bak) != `{"keep":1}` {
		t.Fatalf("backup = %q, %v", bak, err)
	}
}

func TestServersForWriteRejectsNonObject(t *testing.T) {
	if _, err := serversForWrite(map[string]interface{}{"mcpServers": []interface{}{}}, "x"); err == nil {
		t.Fatal("expected error for non-object mcpServers")
	}
	s, err := serversForWrite(map[string]interface{}{}, "x")
	if err != nil || s == nil {
		t.Fatalf("missing mcpServers: %v %v", s, err)
	}
}
