package cmd

import (
	"fmt"
	"os"

	"rsdish/persist" // Import the persist package

	"github.com/spf13/cobra"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Manage collections of media.",
	Long:  `The collect command allows you to add, remove, and list details about your media collections stored in ~/.rsdish.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is given, show help for collect.
		cmd.Help()
	},
}

var collectAddCmd = &cobra.Command{
	Use:   "add <shortname> <uuid>",
	Short: "Add a new collection to ~/.rsdish.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		shortname := args[0]
		uuid := args[1]

		cfg, err := persist.LoadConfig() // Use persist.LoadConfig
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Check if collection with shortname already exists
		for _, col := range cfg.Collections {
			if col.Short == shortname {
				fmt.Fprintf(os.Stderr, "Error: Collection with shortname '%s' already exists.\n", shortname)
				os.Exit(1)
			}
		}

		newCollection := persist.Collection{Short: shortname, UUID: uuid} // Use persist.Collection
		cfg.Collections = append(cfg.Collections, newCollection)

		if err := persist.SaveConfig(cfg); err != nil { // Use persist.SaveConfig
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully added collection: Shortname='%s', UUID='%s'\n", shortname, uuid)
	},
}

var collectRemoveCmd = &cobra.Command{
	Use:   "remove <shortname>",
	Short: "Remove an existing collection from ~/.rsdish.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		shortname := args[0]

		cfg, err := persist.LoadConfig() // Use persist.LoadConfig
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		found := false
		var updatedCollections []persist.Collection // Use persist.Collection
		for _, col := range cfg.Collections {
			if col.Short == shortname {
				found = true
			} else {
				updatedCollections = append(updatedCollections, col)
			}
		}

		if !found {
			fmt.Fprintf(os.Stderr, "Error: Collection with shortname '%s' not found.\n", shortname)
			os.Exit(1)
		}

		cfg.Collections = updatedCollections

		if err := persist.SaveConfig(cfg); err != nil { // Use persist.SaveConfig
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully removed collection: Shortname='%s'\n", shortname)
	},
}

// New collectLsCmd for listing collections
var collectLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all collections from ~/.rsdish.",
	Long:  `The 'ls' subcommand lists all currently defined collections and their UUIDs from the ~/.rsdish configuration file.`,
	Args:  cobra.NoArgs, // No arguments expected for 'ls'
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing all collections from ~/.rsdish:")
		cfg, err := persist.LoadConfig() // Use persist.LoadConfig
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if len(cfg.Collections) == 0 {
			fmt.Println("  No collections found.")
		} else {
			for _, col := range cfg.Collections {
				fmt.Printf("  Shortname: %s, UUID: %s\n", col.Short, col.UUID)
			}
		}
	},
}

func init() {
	collectCmd.AddCommand(collectAddCmd)
	collectCmd.AddCommand(collectRemoveCmd)
	collectCmd.AddCommand(collectLsCmd) // Add the new 'ls' subcommand
	// Removed: collectCmd.Flags().BoolP("verbose", "v", false, "List all collections with verbose details from ~/.rsdish")
}
