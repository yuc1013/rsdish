package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"rsdish/logi"
	"rsdish/persist"
	"rsdish/phys"

	"github.com/spf13/cobra"
)

var (
	dropLibraryID string // The UUID or shortname of the library to drop from
)

var dropCmd = &cobra.Command{
	Use:   "drop <Relative FilePath>...",
	Short: "Generate a script to delete a file from all volumes of a library.",
	Long: `The drop command generates a safe Rclone script to delete specified files
from all volumes (buffers and storages) of a given library.

This command does not delete files directly. It generates a script that you
must review and run manually to perform the deletion. The generated script
will use 'rclone delete' with the '--dry-run' flag for safety. You must
manually remove this flag to execute the actual deletion.

You can specify multiple file paths.

Examples:
  rsdish drop "photos/2025/vacation.jpg" --from my_photo_archive
  rsdish drop "videos/A.mp4" "videos/B.mp4" --from <UUID>
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Build Physical and Logical Trees
		phys.BuildPhysTree()
		logi.BuildLogiTree()

		if dropLibraryID == "" {
			log.Fatal("Error: The '--from' flag is required to specify the library.")
		}

		// 2. Resolve Library ID from shortname using persist.LoadConfig
		resolvedUUID := dropLibraryID // 默认假设输入已经是 UUID
		cfg, err := persist.LoadConfig()
		if err != nil {
			log.Fatalf("Error: Failed to load user config: %v", err)
		}

		for _, collection := range cfg.Collections {
			if collection.Short == dropLibraryID {
				resolvedUUID = collection.UUID
				log.Printf("Resolved shortname '%s' to UUID '%s'", dropLibraryID, resolvedUUID)
				break
			}
		}

		// 检查解析出的 UUID 是否存在于逻辑树中
		library, ok := logi.LogiTree[resolvedUUID]
		if !ok {
			log.Fatalf("Error: Library with ID '%s' (resolved from '%s') not found in LogiTree. Please check your volume configurations.", resolvedUUID, dropLibraryID)
		}

		// 3. Generate Rclone 'delete' Commands
		var rcloneCmds strings.Builder

		// Add safety note to the script
		rcloneCmds.WriteString("# This script will delete files from the following library volumes.\n")
		rcloneCmds.WriteString("# For safety, it is generated with the '--dry-run' flag.\n")
		rcloneCmds.WriteString("# To perform the actual deletion, please review this script and remove the '--dry-run' flag.\n\n")

		// Add OS-specific header
		if runtime.GOOS == "windows" {
			rcloneCmds.WriteString("@echo off\n\n")
		} else {
			rcloneCmds.WriteString("#!/bin/bash\nset -e\n\n")
		}

		allVolumes := append(library.Buffers, library.Storages...)
		if len(allVolumes) == 0 {
			log.Fatalf("Error: Library '%s' has no volumes to drop files from.", resolvedUUID)
		}

		// Generate a delete command for each file for each volume
		for _, volume := range allVolumes {
			for _, relativePath := range args {
				// Construct the full path to the file to be deleted
				fullPath := filepath.Join(volume.BasePath, relativePath)

				// Use rclone's delete command. For safety, we use '--dry-run'
				cmdStr := fmt.Sprintf("rclone delete \"%s\" --dry-run", fullPath)
				rcloneCmds.WriteString(cmdStr + "\n")
			}
			rcloneCmds.WriteString("\n") // Add a newline between volumes for readability
		}

		// 4. Write Script to File
		suffix := ".sh"
		if runtime.GOOS == "windows" {
			suffix = ".bat"
		}
		scriptFileName := fmt.Sprintf("rsdish_drop_%s%s", resolvedUUID[:8], suffix)

		fileMode := os.FileMode(0644)
		if runtime.GOOS != "windows" {
			fileMode = 0755
		}

		err = os.WriteFile(scriptFileName, []byte(rcloneCmds.String()), fileMode)
		if err != nil {
			log.Fatalf("Error writing drop script to file '%s': %v", scriptFileName, err)
		}

		fmt.Printf("\nSuccessfully generated deletion script: %s\n", scriptFileName)
		fmt.Println("Please review the script before running it.")
	},
}

func init() {
	rootCmd.AddCommand(dropCmd)
	dropCmd.Flags().StringVar(&dropLibraryID, "from", "", "Required: UUID or shortname of the library to delete files from.")
	dropCmd.MarkFlagRequired("from")
}
