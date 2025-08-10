package cmd

import (
	"fmt"
	"log"
	"rsdish/logi"
	"rsdish/persist"
	"rsdish/phys"

	"github.com/spf13/cobra"
)

var (
	linkAll       bool   // Flag to indicate all libraries should be linked
	linkDryRun    bool   // Flag to enable dry-run mode
	linkLibraryID string // A specific library's UUID or shortname, provided as an argument
)

var linkCmd = &cobra.Command{
	Use:   "link <UUID|shortname>",
	Short: "Create logical links within volumes.",
	Long: `The link command creates the logical links (symlinks or cheatfiles) 
within the configured volumes based on the 'link_creat' setting in their
volume.toml files.

You can specify a single library to process or use the --all flag for all libraries.
The --dry-run flag can be used to preview the operations without making any changes.

Examples:
  rsdish link <uuid_or_shortname>
  rsdish link --all
  rsdish link <uuid_or_shortname> --dry-run
  rsdish link --all --dry-run`,
	Args: func(cmd *cobra.Command, args []string) error {
		// Custom validation to handle the --all flag and positional arguments.
		// A single positional argument is accepted, or the --all flag, but not both.
		if len(args) > 1 {
			return fmt.Errorf("accepts at most one argument, received %d", len(args))
		}
		if len(args) == 1 && linkAll {
			return fmt.Errorf("cannot use both a library ID and the --all flag")
		}
		if len(args) == 0 && !linkAll {
			return fmt.Errorf("must specify a library ID or use the --all flag")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Build Physical and Logical Trees
		phys.BuildPhysTree()
		logi.BuildLogiTree()

		// 2. Resolve Library ID
		if len(args) > 0 {
			linkLibraryID = args[0]
		}

		resolvedUUID := ""
		if linkLibraryID != "" {
			var err error
			resolvedUUID, err = persist.ResolveCollectionID(linkLibraryID)
			if err != nil {
				log.Printf("Warning: Could not resolve '%s': %v. Attempting to use it directly as a UUID.", linkLibraryID, err)
				resolvedUUID = linkLibraryID // Fallback to using it as is
			}

			if _, exists := logi.LogiTree[resolvedUUID]; !exists {
				log.Fatalf("Error: Library with ID '%s' (resolved to '%s') not found in LogiTree. Please check your volume configurations.", linkLibraryID, resolvedUUID)
			}
		}

		// 3. Execute Link Operation
		if linkDryRun {
			log.Println("--- DRY RUN MODE: No changes will be made to the filesystem. ---")
		}

		if linkAll {
			log.Println("Starting link operation for ALL libraries...")
			err := logi.LinkAllLibrary(linkDryRun)
			if err != nil {
				log.Fatalf("Error during link operation for all libraries: %v", err)
			}
		} else {
			if resolvedUUID == "" {
				log.Fatal("Error: No library ID specified.")
			}
			log.Printf("Starting link operation for library: %s\n", resolvedUUID)
			err := logi.LinkLibrary(resolvedUUID, linkDryRun)
			if err != nil {
				log.Fatalf("Error during link operation for library '%s': %v", resolvedUUID, err)
			}
		}

		log.Println("Link operation completed.")
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)

	linkCmd.Flags().BoolVar(&linkAll, "all", false, "Process all configured libraries.")
	linkCmd.Flags().BoolVar(&linkDryRun, "dry-run", false, "Simulate the link creation process without making any changes to the filesystem.")
}
