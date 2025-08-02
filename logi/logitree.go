package logi

import (
	"log"
	"rsdish/persist" // To access persist.VolumeConfig
	"rsdish/phys"    // To access phys.PhysTree
)

// LogiTree is a placeholder for the logical representation of all libraries and volumes.
// This would be populated by a function like `BuildLogiTree` which processes `phys.PhysTree`.
var LogiTree = make(map[string]*Library)

// Library represents a logical collection of volumes.
type Library struct {
	UUID     string
	Buffers  []*Volume // Volumes with mode="buffer"
	Storages []*Volume // Volumes with mode="storage"
}

// Volume represents a single logical volume, derived from a physical volume.
type Volume struct {
	UUID     string
	Mode     string
	BasePath string
	Config   *persist.VolumeConfig // Stores the full parsed volume.toml config
}

// BuildLogiTree processes the physical volume tree (phys.PhysTree)
// and constructs the logical library tree (LogiTree).
// It groups volumes by their library UUID and categorizes them as buffers or storages.
//
// NOTE: phys.PhysTree is assumed to be read-only after its generation and
// is regenerated on each command execution, thus no explicit locking is required here.
func BuildLogiTree() {
	// Clear the LogiTree before rebuilding to ensure a fresh state
	LogiTree = make(map[string]*Library)

	log.Println("Building logical library tree from physical volumes...")

	// Iterate over all physical volumes discovered by phys.BuildPhysTree
	// No explicit locking on phys.PhysTree is needed as it's treated as read-only after initial build.
	if len(phys.PhysTree) == 0 {
		log.Println("No physical volumes found to build logical tree. Ensure 'rsdish scan mp' and 'rsdish scan lib' run first.")
		return
	}

	for basePath, volConfig := range phys.PhysTree {
		libraryUUID := volConfig.Library.UUID
		volumeMode := volConfig.Volume.Mode

		// If the library doesn't exist in LogiTree yet, create it
		if _, exists := LogiTree[libraryUUID]; !exists {
			LogiTree[libraryUUID] = &Library{
				UUID:     libraryUUID,
				Buffers:  []*Volume{},
				Storages: []*Volume{},
			}
		}

		// Create a new logical Volume object
		logicalVolume := &Volume{
			UUID:     libraryUUID, // The library UUID this volume belongs to
			Mode:     volumeMode,
			BasePath: basePath, // The base path of the volume.toml (its parent directory)
			Config:   volConfig,
		}

		// Add the logical volume to the appropriate slice within its library
		switch volumeMode {
		case "buffer":
			LogiTree[libraryUUID].Buffers = append(LogiTree[libraryUUID].Buffers, logicalVolume)
			log.Printf("  Added buffer volume '%s' to library '%s'", basePath, libraryUUID)
		case "storage":
			LogiTree[libraryUUID].Storages = append(LogiTree[libraryUUID].Storages, logicalVolume)
			log.Printf("  Added storage volume '%s' to library '%s'", basePath, libraryUUID)
		default:
			// This case should ideally not be hit if phys.validateVolumeConfig is robust
			log.Printf("  Warning: Volume '%s' has unknown mode '%s'. Skipping.", basePath, volumeMode)
		}
	}

	log.Printf("Finished building logical library tree. Found %d libraries.", len(LogiTree))
}
