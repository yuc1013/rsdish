package cmd

import (
	"fmt"
	"log"
	"rsdish/logi"    // Import the logi package
	"rsdish/persist" // Import persist for collection resolution (if needed by other subcommands)
	"rsdish/phys"    // Import the phys package

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan and display system information related to rsdish volumes.",
	Long:  `The scan command provides tools to inspect discovered mount points and configured volumes.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var scanLibCmd = &cobra.Command{
	Use:   "lib",
	Short: "Scan and display configured libraries and their volumes.",
	Long:  `Scans the system for 'volume.toml' files, builds the physical and logical volume trees, and displays the discovered libraries and their associated volumes.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Build the Physical Tree first
		phys.BuildPhysTree()

		// 2. Then, build the Logical Tree from the Physical Tree
		logi.BuildLogiTree()

		// Display the organized LogiTree
		if len(logi.LogiTree) == 0 {
			fmt.Println("No libraries found. Ensure volume.toml files are correctly placed.")
			return
		}

		fmt.Println("\n--- Discovered Libraries and Volumes ---")
		for uuid, library := range logi.LogiTree {
			fmt.Printf("Library UUID: %s\n", uuid)
			fmt.Printf("  Buffers (%d):\n", len(library.Buffers))
			if len(library.Buffers) == 0 {
				fmt.Println("    (None)")
			}
			for _, vol := range library.Buffers {
				fmt.Printf("    - Path: %s (UUID: %s)\n", vol.BasePath, vol.UUID)
				if vol.Config != nil && vol.Config.Volume.Note != "" {
					fmt.Printf("      Note: %s\n", vol.Config.Volume.Note)
				}
				if vol.Config != nil && vol.Config.Advanced.RcloneArguments != "" {
					fmt.Printf("      Rclone Args: \"%s\"\n", vol.Config.Advanced.RcloneArguments)
				}
			}

			fmt.Printf("  Storages (%d):\n", len(library.Storages))
			if len(library.Storages) == 0 {
				fmt.Println("    (None)")
			}
			for _, vol := range library.Storages {
				fmt.Printf("    - Path: %s (UUID: %s)\n", vol.BasePath, vol.UUID)
				if vol.Config != nil && vol.Config.Volume.Note != "" {
					fmt.Printf("      Note: %s\n", vol.Config.Volume.Note)
				}
				if vol.Config != nil && vol.Config.Advanced.RcloneArguments != "" {
					fmt.Printf("      Rclone Args: \"%s\"\n", vol.Config.Advanced.RcloneArguments)
				}
			}
			fmt.Println("") // Add a newline for separation
		}
	},
}

var scanMpCmd = &cobra.Command{
	Use:   "mp",
	Short: "Scan and display system mount points.",
	Long:  `Scans the system for active mount points and displays them. This can help debug volume discovery issues.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		mountPoints, err := phys.GetMountPoints()
		if err != nil {
			log.Fatalf("Error getting mount points: %v", err)
		}

		cfg, err := persist.LoadConfig()
		if err != nil {
			log.Printf("Warning: Could not load user config for additional mount points: %v", err)
		}

		fmt.Println("--- Discovered Mount Points ---")
		if len(mountPoints) == 0 && (cfg == nil || len(cfg.AdditionalMountpoints) == 0) {
			fmt.Println("No mount points found.")
			return
		}

		allMPs := make(map[string]struct{})
		for _, mp := range mountPoints {
			allMPs[mp] = struct{}{}
		}
		if cfg != nil {
			for _, amp := range cfg.AdditionalMountpoints {
				allMPs[amp] = struct{}{}
			}
		}

		for mp := range allMPs {
			fmt.Printf("- %s\n", mp)
		}
		fmt.Println("")
	},
}

func init() {
	scanCmd.AddCommand(scanLibCmd)
	scanCmd.AddCommand(scanMpCmd)
	rootCmd.AddCommand(scanCmd)
}
