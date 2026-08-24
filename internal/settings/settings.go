// Package settings defines the runtime configuration of a ConfigHub process.
package settings

import (
	"fmt"
	"os"
	"path/filepath"
)

// Settings holds the values needed to start one ConfigHub process.
type Settings struct {
	HTTPAddr       string
	DataDir        string
	NodeID         string
	DefaultQuota   int64
	NamespaceCount int
}

// Default returns a development-friendly settings object.
func Default() Settings {
	return Settings{
		HTTPAddr:       "127.0.0.1:7790",
		DataDir:        filepath.Join(os.TempDir(), "confighub"),
		NodeID:         "node-a",
		DefaultQuota:   1 << 20,
		NamespaceCount: 4,
	}
}

// Validate checks that the settings can produce a runnable cluster.
func (s Settings) Validate() error {
	if s.DataDir == "" {
		return fmt.Errorf("confighub: data dir is required")
	}
	if s.NodeID == "" {
		return fmt.Errorf("confighub: node id is required")
	}
	if s.DefaultQuota < 0 {
		return fmt.Errorf("confighub: quota cannot be negative")
	}
	if s.NamespaceCount < 1 {
		return fmt.Errorf("confighub: namespace count must be positive")
	}
	return nil
}
