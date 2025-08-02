package persist

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
)

// VolumeConfig represents the structure of a volume.toml file.
type VolumeConfig struct {
	Library  LibrarySection  `toml:"library"`
	Volume   VolumeSection   `toml:"volume"`
	Advanced AdvancedSection `toml:"advanced"` // Reintroducing the advanced section
}

// LibrarySection corresponds to the [library] table within VolumeConfig.
type LibrarySection struct {
	UUID string `toml:"uuid"`
}

// VolumeSection corresponds to the [volume] table within VolumeConfig.
type VolumeSection struct {
	Mode string `toml:"mode"`           // REQUIRED FROM: (storage/buffer)
	Note string `toml:"note,omitempty"` // Note can be optional
}

// AdvancedSection corresponds to the [advanced] table within VolumeConfig.
// Fields are marked 'omitempty' because they can be optional in the TOML.
type AdvancedSection struct {
	RcloneArguments string `toml:"rclone_arguments,omitempty"` // Now optional in TOML
	LinkCreat       string `toml:"link_creat,omitempty"`       // Now optional in TOML
}

// SaveTomlConfig writes any TOML-serializable struct to the specified path.
func SaveTomlConfig(data interface{}, outputPath string) error {
	marshaledData, err := toml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal config to TOML: %w", err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
	}

	tmpFileName := filepath.Base(outputPath) + ".tmp"
	tmpFile, err := os.CreateTemp(outputDir, tmpFileName)
	if err != nil {
		return fmt.Errorf("failed to create temporary config file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(marshaledData); err != nil {
		return fmt.Errorf("failed to write to temporary config file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary config file: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), outputPath); err != nil {
		return fmt.Errorf("failed to rename temporary config file to '%s': %w", outputPath, err)
	}

	if err := os.Chmod(outputPath, 0644); err != nil {
		return fmt.Errorf("failed to set permissions for config file '%s': %w", outputPath, err)
	}

	return nil
}

// ResolveCollectionID takes an ID (shortname or UUID) and returns its corresponding UUID.
func ResolveCollectionID(id string) (string, error) {
	if _, err := uuid.Parse(id); err == nil {
		return id, nil // It's already a UUID, no resolution needed
	}

	cfg, err := LoadConfig() // Uses LoadConfig from persist/user.go
	if err != nil {
		return id, fmt.Errorf("could not load user config to resolve shortname '%s': %w", id, err)
	}

	for _, col := range cfg.Collections {
		if col.Short == id {
			return col.UUID, nil // Found a matching shortname, return its UUID
		}
	}

	return id, nil
}
