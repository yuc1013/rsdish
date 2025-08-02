package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rsdish",
	Short: "RSDish is a tool for managing your media libraries.",
	Long: `A comprehensive CLI tool for organizing, scanning,
syncing, and managing your media collections.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default action if no subcommand is given
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add global flags here if needed
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rsdish.yaml)")

	// Add subcommands
	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(dropCmd)
}
