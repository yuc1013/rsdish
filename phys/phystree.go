package phys

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"rsdish/persist"

	"github.com/BurntSushi/toml"
)

// PhysTree maps a volume's mount point path (e.g., /mnt/my-drive/volumes/volumeA)
// to its parsed VolumeConfig.
var (
	PhysTree = make(map[string]*persist.VolumeConfig)
	mu       sync.Mutex
)

// BuildPhysTree orchestrates the discovery and loading of all configured volumes.
func BuildPhysTree() {
	mu.Lock()
	PhysTree = make(map[string]*persist.VolumeConfig) // Re-initialize the map
	mu.Unlock()

	mps, err := getAllMountpointsIncludeAdditionals()
	if err != nil {
		log.Printf("Failed to get mountpoints: %v", err)
		return
	}

	var wg sync.WaitGroup
	for _, mp := range mps {
		wg.Add(1)
		go func(mountpoint string) {
			defer wg.Done()
			if err := LoadTomlFromMountpoint(mountpoint); err != nil {
				log.Printf("Failed to process TOML from mountpoint %s: %v", mountpoint, err)
			}
		}(mp)
	}
	wg.Wait()
}

// getAllMountpointsIncludeAdditionals combines system mount points with user-defined
// additional mount points from ~/.rsdish and returns a unique list of paths.
func getAllMountpointsIncludeAdditionals() ([]string, error) {
	uniqueMounts := make(map[string]struct{})

	systemMps, err := GetMountPoints()
	if err != nil {
		return nil, fmt.Errorf("failed to get system mount points: %w", err)
	}
	for _, mp := range systemMps {
		uniqueMounts[mp] = struct{}{}
	}

	cfg, err := persist.LoadConfig()
	if err != nil {
		log.Printf("Warning: Failed to load user config for additional mount points: %v", err)
	} else {
		for _, amp := range cfg.AdditionalMountpoints {
			uniqueMounts[amp] = struct{}{}
		}
	}

	var result []string
	for mp := range uniqueMounts {
		result = append(result, mp)
	}

	return result, nil
}

// LoadTomlFromMountpoint now walks the 'volumes' subdirectory within the given mountpoint
// to find and load 'volume.toml' files.
func LoadTomlFromMountpoint(mp string) error {
	volumesDirPath := filepath.Join(mp, "volumes")

	if _, err := os.Stat(volumesDirPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to stat 'volumes' directory at '%s': %w", volumesDirPath, err)
	}

	err := fs.WalkDir(os.DirFS(volumesDirPath), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("Error walking path %s within %s/volumes: %v", path, mp, err)
			return nil
		}

		if d.Type().IsRegular() && strings.ToLower(d.Name()) == "volume.toml" {
			fullTomlPath := filepath.Join(volumesDirPath, path)

			var volumeCfg persist.VolumeConfig
			_, decodeErr := toml.DecodeFile(fullTomlPath, &volumeCfg)
			if decodeErr != nil {
				log.Printf("Failed to decode TOML file '%s': %v", fullTomlPath, decodeErr)
				return nil
			}

			// Validate the loaded volume configuration
			if validationErr := validateVolumeConfig(&volumeCfg); validationErr != nil {
				log.Printf("Validation failed for '%s': %v", fullTomlPath, validationErr)
				return nil
			}

			volumeBasePath := filepath.Dir(fullTomlPath)

			mu.Lock()
			PhysTree[volumeBasePath] = &volumeCfg
			mu.Unlock()

			log.Printf("Successfully loaded volume from %s", volumeBasePath)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk 'volumes' directory in '%s': %w", mp, err)
	}

	return nil
}

// validateVolumeConfig checks if the loaded VolumeConfig meets required criteria.
// It enforces that 'library.uuid' and 'volume.mode' must be present and valid.
// Other fields like 'note', 'rclone_arguments', and 'link_creat' are optional.
func validateVolumeConfig(cfg *persist.VolumeConfig) error {
	// 1. Validate 'library.uuid' (Required)
	if cfg.Library.UUID == "" {
		return fmt.Errorf("volume config missing required 'library.uuid'")
	}

	// 2. Validate 'volume.mode' (Required and must be specific values)
	switch cfg.Volume.Mode {
	case "storage", "buffer":
		// Valid modes
	case "":
		return fmt.Errorf("volume config missing required 'volume.mode'")
	default:
		return fmt.Errorf("volume config has invalid 'volume.mode': '%s'. Must be 'storage' or 'buffer'", cfg.Volume.Mode)
	}

	// 3. Validate 'volume.note' (Optional)
	// No specific validation needed as it's 'ANY' and omitempty.

	// 4. Validate 'advanced.rclone_arguments' (Optional)
	// No specific validation needed as it's an example string and optional.

	// 5. Validate 'advanced.link_creat' (Optional, but if present, must be specific values)
	if cfg.Advanced.LinkCreat != "" { // Only validate if the field is present/not empty
		switch cfg.Advanced.LinkCreat {
		case "none", "symlink", "cheatfile":
			// Valid link creation types
		default:
			return fmt.Errorf("volume config has invalid 'advanced.link_creat': '%s'. Must be 'none', 'symlink', or 'cheatfile'", cfg.Advanced.LinkCreat)
		}
	}
	// If cfg.Advanced.LinkCreat is empty, it's considered valid because it's optional.

	return nil
}
