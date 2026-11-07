// Package transfer handles portable profile archives and directory copies.
package transfer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Counts struct {
	Files int   `json:"files"`
	Bytes int64 `json:"bytes"`
	// Browser is the browser ID recorded in the archive's bpm-export.json,
	// or empty when the source has no readable metadata.
	Browser string `json:"-"`
}

// maxMetadataSize bounds how much of bpm-export.json is read.
const maxMetadataSize = 64 << 10

var excluded = map[string]bool{"Cache": true, "Code Cache": true, "GPUCache": true, "DawnCache": true, "DawnGraphiteCache": true, "DawnWebGPUCache": true, "ShaderCache": true, "GrShaderCache": true, "GraphiteDawnCache": true, "Crashpad": true, "Crash Reports": true, "BrowserMetrics": true, "DevToolsActivePort": true, ".bpm.lock": true, ".bpm.lock.guard": true, "component_crx_cache": true, "optimization_guide_hint_cache_store": true}

func skip(name string) bool { return excluded[name] || strings.HasPrefix(name, "Singleton") }
func format(s string) string {
	s = strings.ToLower(s)
	if strings.HasSuffix(s, ".tar.gz") || strings.HasSuffix(s, ".tgz") {
		return "tar"
	}
	if strings.HasSuffix(s, ".zip") {
		return "zip"
	}
	return ""
}
func Export(src, dst, name, browser string) (c Counts, err error) {
	kind := format(dst)
	if kind == "" {
		return c, fmt.Errorf("unsupported destination extension: use .tar.gz, .tgz, or .zip")
	}
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return c, err
	}
	defer func() {
		if e := f.Close(); err == nil {
			err = e
		}
		if err != nil {
			os.Remove(dst)
		}
	}()
	var add func(string, os.FileInfo, io.Reader) error
	var closeArchive func() error
	if kind == "zip" {
		w := zip.NewWriter(f)
		add = func(n string, info os.FileInfo, r io.Reader) error {
			h, e := zip.FileInfoHeader(info)
			if e != nil {
				return e
			}
			h.Name = n
			if info.IsDir() {
				h.Name += "/"
			} else {
				h.Method = zip.Deflate
			}
			out, e := w.CreateHeader(h)
			if e != nil {
				return e
			}
			if r != nil {
				_, e = io.Copy(out, r)
			}
			return e
		}
		closeArchive = w.Close
	} else {
		gz := gzip.NewWriter(f)
		w := tar.NewWriter(gz)
		add = func(n string, info os.FileInfo, r io.Reader) error {
			h, e := tar.FileInfoHeader(info, "")
			if e != nil {
				return e
			}
			h.Name = n
			if e = w.WriteHeader(h); e != nil {
				return e
			}
			if r != nil {
				_, e = io.Copy(w, r)
			}
			return e
		}
		closeArchive = func() error {
			e := w.Close()
			e2 := gz.Close()
			if e != nil {
				return e
			}
			return e2
		}
	}
	err = filepath.Walk(src, func(p string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if p == src {
			return nil
		}
		if skip(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return nil
		}
		rel, e := filepath.Rel(src, p)
		if e != nil {
			return e
		}
		if filepath.ToSlash(rel) == "bpm-export.json" {
			return nil
		}
		var r *os.File
		if !info.IsDir() {
			r, e = os.Open(p)
			if e != nil {
				return e
			}
			defer r.Close()
			c.Files++
			c.Bytes += info.Size()
		}
		if r == nil {
			return add(filepath.ToSlash(rel), info, nil)
		}
		return add(filepath.ToSlash(rel), info, r)
	})
	if err == nil {
		data, _ := json.Marshal(map[string]any{"bpm_export_version": 1, "profile": name, "browser": browser, "created_at": time.Now().UTC(), "source_os": runtime.GOOS})
		info := metadataInfo{int64(len(data))}
		err = add("bpm-export.json", info, strings.NewReader(string(data)))
	}
	if e := closeArchive(); err == nil {
		err = e
	}
	return c, err
}

type metadataInfo struct{ size int64 }

func (i metadataInfo) Name() string       { return "bpm-export.json" }
func (i metadataInfo) Size() int64        { return i.size }
func (i metadataInfo) Mode() os.FileMode  { return 0600 }
func (i metadataInfo) ModTime() time.Time { return time.Now() }
func (i metadataInfo) IsDir() bool        { return false }
func (i metadataInfo) Sys() any           { return nil }
func safe(n string) (string, error) {
	n = strings.ReplaceAll(n, `\`, `/`)
	if strings.HasPrefix(n, "/") || (len(n) >= 2 && n[1] == ':') {
		return "", fmt.Errorf("unsafe archive path %q", n)
	}
	for _, p := range strings.Split(n, "/") {
		if p == ".." {
			return "", fmt.Errorf("unsafe archive path %q", n)
		}
	}
	return path.Clean(n), nil
}
func write(dst, n string, dir bool, r io.Reader, c *Counts) error {
	n, e := safe(n)
	if e != nil {
		return e
	}
	if n == "." {
		return nil
	}
	for _, part := range strings.Split(n, "/") {
		if skip(part) {
			return nil
		}
	}
	if n == "bpm-export.json" {
		if !dir && r != nil {
			var meta struct {
				Browser string `json:"browser"`
			}
			if json.NewDecoder(io.LimitReader(r, maxMetadataSize)).Decode(&meta) == nil {
				c.Browser = meta.Browser
			}
		}
		return nil
	}
	p := filepath.Join(dst, filepath.FromSlash(n))
	if dir {
		return os.MkdirAll(p, 0700)
	}
	if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		return e
	}
	f, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	size, e := io.Copy(f, r)
	ce := f.Close()
	if e != nil {
		return e
	}
	if ce != nil {
		return ce
	}
	c.Files++
	c.Bytes += size
	return nil
}
func Unpack(src, dst string) (c Counts, err error) {
	info, err := os.Stat(src)
	if err != nil {
		return c, err
	}
	if info.IsDir() {
		err = filepath.Walk(src, func(p string, i os.FileInfo, e error) error {
			if e != nil {
				return e
			}
			if p == src {
				return nil
			}
			if skip(i.Name()) {
				if i.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if i.Mode()&os.ModeSymlink != 0 {
				return nil
			}
			if !i.IsDir() && !i.Mode().IsRegular() {
				return nil
			}
			rel, e := filepath.Rel(src, p)
			if e != nil {
				return e
			}
			if i.IsDir() {
				return write(dst, filepath.ToSlash(rel), true, nil, &c)
			}
			f, e := os.Open(p)
			if e != nil {
				return e
			}
			defer f.Close()
			return write(dst, filepath.ToSlash(rel), false, f, &c)
		})
		return c, err
	}
	switch format(src) {
	case "zip":
		z, e := zip.OpenReader(src)
		if e != nil {
			return c, e
		}
		defer z.Close()
		for _, f := range z.File {
			if _, e = safe(f.Name); e != nil {
				return c, e
			}
			if f.Mode()&os.ModeSymlink != 0 || (!f.FileInfo().IsDir() && !f.Mode().IsRegular()) {
				return c, fmt.Errorf("unsupported archive entry %q", f.Name)
			}
			if f.FileInfo().IsDir() {
				e = write(dst, f.Name, true, nil, &c)
			} else {
				var r io.ReadCloser
				r, e = f.Open()
				if e == nil {
					e = write(dst, f.Name, false, r, &c)
					r.Close()
				}
			}
			if e != nil {
				return c, e
			}
		}
	case "tar":
		f, e := os.Open(src)
		if e != nil {
			return c, e
		}
		defer f.Close()
		gz, e := gzip.NewReader(f)
		if e != nil {
			return c, e
		}
		defer gz.Close()
		t := tar.NewReader(gz)
		for {
			h, e := t.Next()
			if e == io.EOF {
				break
			}
			if e != nil {
				return c, e
			}
			if _, e = safe(h.Name); e != nil {
				return c, e
			}
			switch h.Typeflag {
			case tar.TypeDir:
				e = write(dst, h.Name, true, nil, &c)
			case tar.TypeReg, tar.TypeRegA:
				e = write(dst, h.Name, false, t, &c)
			default:
				return c, fmt.Errorf("unsupported archive entry %q", h.Name)
			}
			if e != nil {
				return c, e
			}
		}
	default:
		return c, fmt.Errorf("unsupported source extension: use .tar.gz, .tgz, or .zip")
	}
	return c, nil
}
func Normalize(dir string) error {
	if _, e := os.Stat(filepath.Join(dir, "Local State")); e == nil {
		return nil
	}
	entries, e := os.ReadDir(dir)
	if e != nil {
		return e
	}
	for _, v := range entries {
		if v.IsDir() && (v.Name() == "Default" || strings.HasPrefix(v.Name(), "Profile ")) {
			return nil
		}
	}
	single := false
	for _, n := range []string{"Cookies", "Login Data", "Preferences", "History"} {
		if _, e = os.Stat(filepath.Join(dir, n)); e == nil {
			single = true
			break
		}
	}
	if !single {
		return fmt.Errorf("source is not a recognized Chrome profile layout")
	}
	tmp, e := os.MkdirTemp(filepath.Dir(dir), ".bpm-layout-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(tmp)
	for _, v := range entries {
		if e = os.Rename(filepath.Join(dir, v.Name()), filepath.Join(tmp, v.Name())); e != nil {
			return e
		}
	}
	if e = os.Rename(tmp, filepath.Join(dir, "Default")); e != nil {
		return e
	}
	return os.WriteFile(filepath.Join(dir, "Local State"), []byte(`{"profile":{"info_cache":{"Default":{"name":"Default"}}}}`), 0600)
}
