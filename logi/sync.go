package logi

import (
	"log"
	"rsdish/persist"
)

// BuildRcloneCmdsForCopy is a helper to build a single Rclone copy command.
// The rclone arguments for this specific copy operation are taken from the 'dstVol's configuration.
func BuildRcloneCmdsForCopy(srcVol *Volume, dstVol *Volume) string {
	// Do not copy a volume to itself
	if srcVol.BasePath == dstVol.BasePath {
		return "" // Return empty string if source and destination are the same
	}

	// Get the rclone arguments from the destination volume's config, as per requirement.
	rcloneArgs := dstVol.Config.Advanced.RcloneArguments

	// Construct the RcloneOptions
	options := persist.RcloneOptions{
		Src:             srcVol.BasePath,
		Dst:             dstVol.BasePath,
		RcloneArguments: rcloneArgs, // Pass rcloneArgs directly, "copy" is handled by persist.BuildRcloneCommand
	}

	// Build the command string
	return persist.BuildRcloneCommand(options)
}

//---

// BuildAppendAllLibrary builds Rclone commands for all libraries.
// It generates `rclone copy` commands to move content from each buffer volume
// to all storage volumes within the same library.
func BuildAppendAllLibrary() []string {
	var allCmds []string
	for _, library := range LogiTree {
		cmds := BuildAppend(library.UUID)
		allCmds = append(allCmds, cmds...)
	}
	return allCmds
}

// BuildAppend builds Rclone copy commands for a specific library's buffer volumes.
// These commands will copy each buffer's content to all storage volumes in the library.
func BuildAppend(uuid string) []string {
	library, ok := LogiTree[uuid]
	if !ok {
		log.Printf("Warning: Library with UUID '%s' not found in LogiTree.", uuid)
		return []string{}
	}

	if len(library.Buffers) == 0 || len(library.Storages) == 0 {
		log.Printf("Library '%s' is missing buffer or storage volumes. Skipping append commands.", uuid)
		return []string{}
	}

	var cmds []string
	for _, bufferVol := range library.Buffers {
		for _, storageVol := range library.Storages {
			cmd := BuildRcloneCmdsForCopy(bufferVol, storageVol)
			if cmd != "" {
				cmds = append(cmds, cmd)
			}
		}
	}
	return cmds
}

//---

// BuildSyncAllLibrary builds Rclone commands for all libraries to synchronize
// content between their storage volumes using bidirectional copy.
func BuildSyncAllLibrary() []string {
	var allCmds []string
	for _, library := range LogiTree {
		cmds := BuildSync(library.UUID)
		allCmds = append(allCmds, cmds...)
	}
	return allCmds
}

// BuildSync builds Rclone copy commands for a specific library.
// These commands will generate `rclone copy` commands between each unique pair
// of storage volumes in the library, in both directions, using the destination's
// rclone_arguments. This avoids redundant command generation.
func BuildSync(uuid string) []string {
	library, ok := LogiTree[uuid]
	if !ok {
		log.Printf("Warning: Library with UUID '%s' not found in LogiTree.", uuid)
		return []string{}
	}

	if len(library.Storages) < 2 {
		log.Printf("Library '%s' needs at least 2 storage volumes for synchronization. Skipping sync commands.", uuid)
		return []string{}
	}

	var cmds []string
	storages := library.Storages

	// Iterate over unique pairs (i, j) where i < j to avoid redundant pairs like (2,4) and (4,2)
	for i := 0; i < len(storages); i++ {
		vol1 := storages[i]
		for j := i + 1; j < len(storages); j++ { // Start j from i+1 to get unique pairs
			vol2 := storages[j]

			// Command 1: Copy from vol1 to vol2, apply vol2's rclone_arguments
			cmd1 := BuildRcloneCmdsForCopy(vol1, vol2)
			if cmd1 != "" {
				cmds = append(cmds, cmd1)
			}

			// Command 2: Copy from vol2 to vol1, apply vol1's rclone_arguments
			cmd2 := BuildRcloneCmdsForCopy(vol2, vol1)
			if cmd2 != "" {
				cmds = append(cmds, cmd2)
			}
		}
	}

	return cmds
}
