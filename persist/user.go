package persist

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml" // Using BurntSushi's TOML parser
)

// Config represents the structure of the ~/.rsdish TOML file
type Config struct {
	Collections           []Collection `toml:"collect"`
	AdditionalMountpoints []string     `toml:"additional_mountpoints"` // Added this field
}

// Collection represents a single collection entry in the TOML file
type Collection struct {
	Short string `toml:"short"`
	UUID  string `toml:"uuid"`
}

const (
	configFileName = ".rsdish"
)

// GetConfigPath returns the full path to the .rsdish config file in the user's home directory.
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, configFileName), nil
}

// LoadConfig reads and parses the .rsdish TOML file into a Config struct.
// If the file doesn't exist, it returns an empty Config and no error.
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	var cfg Config
	// Use toml.DecodeFile which handles file reading and unmarshaling directly.
	_, err = toml.DecodeFile(configPath, &cfg)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, return an empty config without error
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to decode config file '%s': %w", configPath, err)
	}
	return &cfg, nil
}

// SaveConfig writes the Config struct back to the .rsdish TOML file.
func SaveConfig(cfg *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create a temporary file to write to first, then rename for atomicity
	tmpFile, err := os.CreateTemp(filepath.Dir(configPath), configFileName+".tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary config file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up temp file on exit
	defer tmpFile.Close()

	encoder := toml.NewEncoder(tmpFile)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config to TOML: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary config file: %w", err)
	}

	// Atomically replace the old config file with the new one
	if err := os.Rename(tmpFile.Name(), configPath); err != nil {
		return fmt.Errorf("failed to rename temporary config file to '%s': %w", configPath, err)
	}

	// Set file permissions after renaming
	// For simplicity, we'll use 0644 here.
	if err := os.Chmod(configPath, 0644); err != nil {
		return fmt.Errorf("failed to set permissions for config file '%s': %w", configPath, err)
	}

	return nil
}
