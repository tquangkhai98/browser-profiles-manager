// Package fsutil copies profile files. When copying directories it copies
// only regular files and directories, and skips files larger than 500 MiB.
package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// maxCopyFileSize is the largest file CopyDir copies; bigger files are skipped
// and reported on stderr.
const maxCopyFileSize = 500 * 1024 * 1024

// CopyDir copies the directory tree at src into dst.
//
// dst must not exist yet or must be an empty directory; CopyDir refuses to
// merge into a populated destination so stale files (for example an old
// SQLite journal) can't survive next to freshly copied data.
//
// Only directories and regular files are copied. Symlinks are not followed
// (Chromium leaves dangling SingletonLock/SingletonCookie/SingletonSocket
// links in a running or crashed profile), and FIFOs, sockets and devices are
// skipped so the copy can't hang on open. Files larger than 500 MiB are
// skipped. Every skipped path is reported on stderr. If dst lies inside src,
// that subtree is not copied into itself.
func CopyDir(src, dst string) error {
	if err := ensureEmptyDest(dst); err != nil {
		return err
	}
	// Resolve a symlinked source root so the walk descends into it.
	src, err := filepath.EvalSymlinks(src)
	if err != nil {
		return err
	}
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}
	// Resolve symlinks in dst's parent too, so a nested destination reached
	// through a symlinked path is still recognised.
	if parent, err := filepath.EvalSymlinks(filepath.Dir(absDst)); err == nil {
		absDst = filepath.Join(parent, filepath.Base(absDst))
	}
	nestedDst := ""
	if isInside(absSrc, absDst) {
		nestedDst = absDst
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if nestedDst != "" && rel != "." && filepath.Join(absSrc, rel) == nestedDst {
			// Don't copy the destination into itself.
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		mode := info.Mode()
		switch {
		case mode.IsDir():
			return os.MkdirAll(target, mode.Perm()|0700)
		case !mode.IsRegular():
			fmt.Fprintf(os.Stderr, "bpm: skipped non-regular file %s (%s)\n", path, mode.Type())
			return nil
		case info.Size() > maxCopyFileSize:
			fmt.Fprintf(os.Stderr, "bpm: skipped %s: %d bytes exceeds the 500 MiB copy limit\n", path, info.Size())
			return nil
		}
		return CopyFile(path, target)
	})
}

// isInside reports whether dst is strictly inside src.
func isInside(src, dst string) bool {
	rel, err := filepath.Rel(src, dst)
	if err != nil || rel == "." || rel == ".." || filepath.IsAbs(rel) {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// ensureEmptyDest returns an error unless dst is missing or an empty directory.
func ensureEmptyDest(dst string) error {
	info, err := os.Lstat(dst)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("destination %s exists and is not a directory", dst)
	}
	entries, err := os.ReadDir(dst)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("destination %s already exists and is not empty", dst)
	}
	return nil
}

// CopyFile copies a file's content and permissions to dst.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	if copyErr == nil {
		copyErr = out.Chmod(info.Mode().Perm())
	}
	closeErr := out.Close()
	if copyErr != nil {
		os.Remove(dst)
		return copyErr
	}
	if closeErr != nil {
		os.Remove(dst)
		return closeErr
	}
	return nil
}
