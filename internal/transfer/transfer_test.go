package transfer

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func put(t *testing.T, root, name, value string) {
	t.Helper()
	p := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(value), 0600); err != nil {
		t.Fatal(err)
	}
}
func TestRoundTrip(t *testing.T) {
	for _, ext := range []string{".tar.gz", ".tgz", ".zip"} {
		t.Run(ext, func(t *testing.T) {
			root := t.TempDir()
			src := filepath.Join(root, "src")
			put(t, src, "Local State", "state")
			put(t, src, "Default/Network/Cookies", "cookies")
			put(t, src, "Default/Cache/trash", "trash")
			archive := filepath.Join(root, "profile"+ext)
			if _, err := Export(src, archive, "work", "chrome"); err != nil {
				t.Fatal(err)
			}
			dst := filepath.Join(root, "dst")
			if err := os.Mkdir(dst, 0700); err != nil {
				t.Fatal(err)
			}
			if _, err := Unpack(archive, dst); err != nil {
				t.Fatal(err)
			}
			if err := Normalize(dst); err != nil {
				t.Fatal(err)
			}
			for n, want := range map[string]string{"Local State": "state", "Default/Network/Cookies": "cookies"} {
				b, err := os.ReadFile(filepath.Join(dst, n))
				if err != nil || string(b) != want {
					t.Fatalf("%s: %q %v", n, b, err)
				}
			}
			if _, err := os.Stat(filepath.Join(dst, "Default/Cache")); !os.IsNotExist(err) {
				t.Fatal("cache retained")
			}
		})
	}
}
func TestSingleProfile(t *testing.T) {
	dir := t.TempDir()
	put(t, dir, "Cookies", "cookie")
	if err := Normalize(dir); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"Default/Cookies", "Local State"} {
		if _, err := os.Stat(filepath.Join(dir, n)); err != nil {
			t.Fatal(err)
		}
	}
}
func TestTraversal(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "bad.zip")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	entry, err := w.Create("../escape")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = entry.Write([]byte("bad"))
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = Unpack(p, t.TempDir()); err == nil {
		t.Fatal("accepted traversal")
	}
}
