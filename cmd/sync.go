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
	syncLibraryID string // UUID or shortname for the library
	syncMode      string // "append" or "storage" or empty for combined
	outputFile    string // Optional output file for the script
)

// generateScript handles writing the script content to a file, with OS-specific headers.
func generateScript(scriptFileName string, commands []string) error {
	var scriptContent strings.Builder

	// Add OS-specific headers
	switch runtime.GOOS {
	case "windows":
		generateWindowsScriptHeader(&scriptContent)
	default: // Unix-like systems
		generateUnixScriptHeader(&scriptContent)
	}

	// Add commands to the script content
	for _, cmdStr := range commands {
		scriptContent.WriteString(cmdStr)
		scriptContent.WriteString("\n")
	}

	// Set permissions for executable scripts (primarily for Unix-like systems)
	var fileMode os.FileMode = 0644 // Default to readable
	if runtime.GOOS != "windows" {
		fileMode = 0755 // Executable for Unix-like
	}

	err := os.WriteFile(scriptFileName, []byte(scriptContent.String()), fileMode)
	if err != nil {
		return fmt.Errorf("error writing Rclone script to file '%s': %w", scriptFileName, err)
	}
	return nil
}

// generateWindowsScriptHeader adds the appropriate header for a Windows batch script.
func generateWindowsScriptHeader(b *strings.Builder) {
	b.WriteString("@echo off\n")
	b.WriteString("\n")
}

// generateUnixScriptHeader adds the appropriate header for a Unix-like shell script.
func generateUnixScriptHeader(b *strings.Builder) {
	b.WriteString("#!/bin/bash\n")
	b.WriteString("set -e\n") // Exit on error for shell scripts
	b.WriteString("\n")
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Generate and save Rclone synchronization scripts.",
	Long: `The sync command generates Rclone copy scripts based on your configured libraries and volumes.
You can specify a single library to sync or process all discovered libraries.

Modes:
  append  : Copies files from buffer volumes to all storage volumes within a library.
            If no library ID is given, performs this for all libraries.
  storage : Performs bidirectional copies between all storage volumes within a library.
            If no library ID is given, performs this for all libraries.
  (none)  : If no mode is specified, generates two separate scripts: one for 'append'
            and one for 'storage', each named appropriately.

Examples:
  rsdish sync --mode append --library <uuid_or_shortname>
  rsdish sync --mode storage --library <uuid_or_shortname>
  rsdish sync --mode append       (for all libraries)
  rsdish sync --mode storage      (for all libraries)
  rsdish sync                     (generates both append and storage scripts for all libraries)
  rsdish sync --library <uuid_or_shortname> (generates both append and storage for a specific library)
  rsdish sync --mode append -o my_append_script.bat (or .sh)
`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Build Physical and Logical Trees
		phys.BuildPhysTree()
		logi.BuildLogiTree()

		var resolvedUUID string
		if syncLibraryID != "" {
			var err error
			// Resolve shortname to UUID if provided
			resolvedUUID, err = persist.ResolveCollectionID(syncLibraryID)
			if err != nil {
				log.Printf("Warning: Could not resolve '%s': %v. Attempting to use it directly as a UUID.", syncLibraryID, err)
				resolvedUUID = syncLibraryID // Fallback to using it as is
			}

			if _, exists := logi.LogiTree[resolvedUUID]; !exists {
				log.Fatalf("Error: Library with ID '%s' (resolved to '%s') not found in LogiTree. Please check your volume configurations.", syncLibraryID, resolvedUUID)
			}
		}

		// Determine if combined script is needed
		generateCombined := syncMode == ""

		// --- Generate Append Commands ---
		if syncMode == "append" || generateCombined {
			var appendCmds []string
			if resolvedUUID != "" {
				fmt.Printf("Generating 'append' commands for library: %s\n", resolvedUUID)
				appendCmds = logi.BuildAppend(resolvedUUID)
			} else {
				fmt.Println("Generating 'append' commands for all libraries.")
				appendCmds = logi.BuildAppendAllLibrary()
			}

			if len(appendCmds) == 0 {
				log.Println("No 'append' Rclone commands generated. Check configurations.")
			} else {
				appendScriptFileName := getOutputFileName("append", resolvedUUID)
				err := generateScript(appendScriptFileName, appendCmds)
				if err != nil {
					log.Fatalf(err.Error())
				}
				fmt.Printf("Successfully generated 'append' script: %s\n", appendScriptFileName)
			}
			if !generateCombined { // If only append was requested, we're done
				return
			}
		}

		// --- Generate Storage Commands ---
		if syncMode == "storage" || generateCombined {
			var storageCmds []string
			if resolvedUUID != "" {
				fmt.Printf("Generating 'storage' sync commands for library: %s\n", resolvedUUID)
				storageCmds = logi.BuildSync(resolvedUUID)
			} else {
				fmt.Println("Generating 'storage' sync commands for all libraries.")
				storageCmds = logi.BuildSyncAllLibrary()
			}

			if len(storageCmds) == 0 {
				log.Println("No 'storage' Rclone commands generated. Check configurations.")
			} else {
				storageScriptFileName := getOutputFileName("storage", resolvedUUID)
				err := generateScript(storageScriptFileName, storageCmds)
				if err != nil {
					log.Fatalf(err.Error())
				}
				fmt.Printf("Successfully generated 'storage' sync script: %s\n", storageScriptFileName)
			}
			if !generateCombined { // If only storage was requested, we're done
				return
			}
		}

		if !generateCombined && syncMode != "append" && syncMode != "storage" {
			log.Fatalf("Error: Invalid sync mode '%s'. Must be 'append', 'storage', or omitted for combined scripts.", syncMode)
		}

		fmt.Println("\nReview the generated script(s) content before executing.")
	},
}

// getOutputFileName determines the script filename based on mode, library ID, and OS.
func getOutputFileName(mode string, libraryUUID string) string {
	if outputFile != "" {
		// If custom output file is provided, use it directly.
		// Note: User is responsible for extension if -o is used with combined mode.
		if syncMode == "" { // If combined mode, append mode/library_id to custom name
			base := strings.TrimSuffix(outputFile, filepath.Ext(outputFile))
			ext := filepath.Ext(outputFile)
			if ext == "" { // Add default extension if none provided with custom name
				if runtime.GOOS == "windows" {
					ext = ".bat"
				} else {
					ext = ".sh"
				}
			}
			if libraryUUID != "" {
				return fmt.Sprintf("%s_%s_%s%s", base, mode, libraryUUID[:8], ext) // Append short UUID part
			}
			return fmt.Sprintf("%s_%s%s", base, mode, ext)
		}
		return outputFile // If not combined mode, just use the provided output file
	}

	// Default file name based on mode, library ID, and OS
	suffix := ".sh"
	if runtime.GOOS == "windows" {
		suffix = ".bat"
	}

	name := fmt.Sprintf("rsdish_%s", mode)
	if libraryUUID != "" {
		name = fmt.Sprintf("%s_%s", name, libraryUUID[:8]) // Use first 8 chars of UUID for brevity
	}
	return fmt.Sprintf("%s%s", name, suffix)
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringVarP(&syncMode, "mode", "m", "", "Optional: Sync mode ('append' or 'storage'). If omitted, both will be generated.")
	syncCmd.Flags().StringVarP(&syncLibraryID, "library", "l", "", "Optional: UUID or shortname of a specific library to sync. If omitted, all libraries will be processed.")
	syncCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Optional: Specify the base output file name for the script (e.g., 'my_sync'). Defaults to 'rsdish_[mode]_[uuid_prefix].[sh/bat]'.")
}
