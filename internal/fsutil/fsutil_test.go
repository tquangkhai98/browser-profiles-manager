package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDir(t *testing.T) {
	tests := []struct {
		name    string
		missing bool
	}{
		{name: "nested directory and file permissions"},
		{name: "missing source", missing: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := t.TempDir()
			src, dst := filepath.Join(base, "src"), filepath.Join(base, "dst")
			if !tt.missing {
				if err := os.MkdirAll(filepath.Join(src, "nested"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(src, "nested", "file"), []byte("profile data"), 0640); err != nil {
					t.Fatal(err)
				}
			}
			err := CopyDir(src, dst)
			if tt.missing {
				if err == nil {
					t.Fatal("expected error for missing source")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dst, "nested", "file")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != "profile data" {
				t.Fatalf("content = %q", data)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != 0640 {
				t.Fatalf("mode = %v, want 0640", info.Mode().Perm())
			}
		})
	}
}
