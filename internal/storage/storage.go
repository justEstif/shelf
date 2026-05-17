package storage

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Entry represents a file or directory in the storage.
type Entry struct {
	Name    string
	Path    string // relative path from the pages root
	IsDir   bool
	Size    int64
	ModTime string
}

// List returns top-level entries in a directory, sorted (folders first, then files).
func List(dir string) ([]Entry, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			continue
		}
		entries = append(entries, Entry{
			Name:    f.Name(),
			Path:    f.Name(),
			IsDir:   f.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
		})
	}

	sortEntries(entries)
	return entries, nil
}

// Walk returns all entries recursively, with paths relative to root.
// Directory sizes include all descendant files.
func Walk(root string) ([]Entry, error) {
	var entries []Entry
	dirSizes := make(map[string]int64)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}

		if !d.IsDir() {
			// Accumulate size into all parent directories
			dir := filepath.Dir(rel)
			for dir != "." {
				dirSizes[dir] += info.Size()
				dir = filepath.Dir(dir)
			}
		}

		entries = append(entries, Entry{
			Name:    d.Name(),
			Path:    filepath.ToSlash(rel),
			IsDir:   d.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
		})
		return nil
	})

	// Apply computed directory sizes
	if dirSizes != nil {
		for i, e := range entries {
			if e.IsDir {
				if s, ok := dirSizes[e.Path]; ok {
					entries[i].Size = s
				}
			}
		}
	}

	return entries, err
}

// SanitizePath prevents path traversal by ensuring the resolved path
// stays within the base directory.
func SanitizePath(baseDir, requestPath string) (string, error) {
	clean := filepath.Clean(requestPath)
	if clean == "." {
		return baseDir, nil
	}
	full := filepath.Join(baseDir, clean)
	rel, err := filepath.Rel(baseDir, full)
	if err != nil {
		return "", os.ErrPermission
	}
	if strings.HasPrefix(rel, "..") {
		return "", os.ErrPermission
	}
	return full, nil
}

// Exists checks whether a path exists on disk.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Delete removes a file or directory.
func Delete(path string) error {
	return os.RemoveAll(path)
}

// ContentType returns the MIME type for a file based on its extension.
func ContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".md":
		return "text/markdown; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".xml":
		return "application/xml"
	default:
		return "application/octet-stream"
	}
}

// FormatFileSize returns a human-readable file size.
func FormatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// DiskUsage returns the total size of all files under dir.
func DiskUsage(dir string) (int64, error) {
	var total int64
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			total += info.Size()
		}
		return nil
	})
	return total, err
}

func sortEntries(entries []Entry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir // folders first
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
}
