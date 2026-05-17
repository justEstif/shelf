package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Visibility levels for files/folders.
const (
	VisibilityPublic    = "public"
	VisibilityPrivate   = "private"   // admin session required
	VisibilityProtected = "protected" // viewer password required
)

// MetadataStore manages per-path visibility settings.
// Stored as a JSON file in the data directory.
type MetadataStore struct {
	sync.RWMutex
	path string
	data map[string]string // path → visibility
}

// NewMetadataStore creates or loads a metadata store from the given data directory.
func NewMetadataStore(dataDir string) *MetadataStore {
	ms := &MetadataStore{
		path: filepath.Join(dataDir, "metadata.json"),
		data: make(map[string]string),
	}
	ms.load()
	return ms
}

func (ms *MetadataStore) load() {
	raw, err := os.ReadFile(ms.path)
	if err != nil {
		return // file doesn't exist yet, that's fine
	}
	json.Unmarshal(raw, &ms.data)
	if ms.data == nil {
		ms.data = make(map[string]string)
	}
}

func (ms *MetadataStore) save() error {
	raw, err := json.MarshalIndent(ms.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ms.path, raw, 0o600)
}

// GetVisibility returns the visibility for a path.
// It checks exact path first, then walks up parent directories.
// Returns VisibilityPublic if no setting is found.
func (ms *MetadataStore) GetVisibility(path string) string {
	ms.RLock()
	defer ms.RUnlock()

	// Normalize: strip leading slash, trailing slash (unless root)
	p := strings.TrimPrefix(path, "/")
	p = strings.TrimSuffix(p, "/")

	// Exact match
	if v, ok := ms.data[p]; ok {
		return v
	}

	// Walk up parent directories
	for {
		parent := filepath.Dir(p)
		if parent == "." || parent == p {
			break
		}
		if v, ok := ms.data[parent]; ok {
			return v
		}
		p = parent
	}

	return VisibilityPublic
}

// SetVisibility sets the visibility for a path.
func (ms *MetadataStore) SetVisibility(path, visibility string) error {
	ms.Lock()
	defer ms.Unlock()

	p := strings.TrimPrefix(path, "/")
	p = strings.TrimSuffix(p, "/")

	if visibility == VisibilityPublic {
		delete(ms.data, p)
	} else {
		ms.data[p] = visibility
	}

	return ms.save()
}

// GetAll returns all visibility settings (for admin UI).
func (ms *MetadataStore) GetAll() map[string]string {
	ms.RLock()
	defer ms.RUnlock()

	result := make(map[string]string, len(ms.data))
	for k, v := range ms.data {
		result[k] = v
	}
	return result
}
