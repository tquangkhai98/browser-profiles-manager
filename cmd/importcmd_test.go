package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

func TestDeriveName(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"/tmp/My Profile.tar.gz", "my-profile"}, {"x.TGZ", "x"}, {"A__B.zip", "a-b"}, {"/tmp/foo bar/", "foo-bar"}, {"---.zip", ""}} {
		if got := deriveName(tc.in); got != tc.want {
			t.Errorf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
}
func TestImportExistingName(t *testing.T) {
	t.Setenv("BPM_CONFIG_DIR", t.TempDir())
	t.Setenv("BPM_DATA_DIR", t.TempDir())
	if _, err := profile.Create("existing", "chrome"); err != nil {
		t.Fatal(err)
	}
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "Local State"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	err := importCmd.RunE(importCmd, []string{src, "existing"})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("got %v", err)
	}
}

func TestListJSONFreshProfile(t *testing.T) {
	t.Setenv("BPM_CONFIG_DIR", t.TempDir())
	t.Setenv("BPM_DATA_DIR", t.TempDir())
	if _, err := profile.Create("fresh", "chrome"); err != nil {
		t.Fatal(err)
	}
	profiles, err := profile.List()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer
	err = printJSON(profiles)
	_ = writer.Close()
	os.Stdout = old
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	_ = reader.Close()
	var result []map[string]any
	if err = json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0]["running"] != false {
		t.Fatalf("unexpected list: %s", data)
	}
}
