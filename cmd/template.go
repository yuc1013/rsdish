package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"rsdish/persist"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage volume templates.",
	Long:  `The template command provides tools for creating and managing volume.toml templates.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var (
	templateFromArg   string
	templateOutputArg string
)

// GenerateVolumeTemplate creates a new persist.VolumeConfig with default values.
// Updated to reflect the reintroduction of the 'advanced' section.
func GenerateVolumeTemplate(libraryUUID string) *persist.VolumeConfig {
	if libraryUUID == "" {
		libraryUUID = uuid.New().String()
	}

	return &persist.VolumeConfig{
		Library: persist.LibrarySection{
			UUID: libraryUUID,
		},
		Volume: persist.VolumeSection{
			Mode: "storage", // Default mode
			Note: "ANY",     // Default note
		},
		Advanced: persist.AdvancedSection{ // Reintroduce and populate the advanced section
			RcloneArguments: "",
			LinkCreat:       "none", // Default link creation type
		},
	}
}

var templateNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new volume.toml template.",
	Long: `Creates a new volume.toml template file with a generated or specified UUID.

Usage:
  rsdish template new [--output <path/to/volume.toml>] [--from <uuid/shortname>]

Arguments:
  --from <uuid/shortname>  : Optional. Specifies the UUID for the 'library' section.
                             If a shortname is provided, it will be resolved to its UUID from ~/.rsdish.
                             If omitted, a new random UUID will be generated.
  --output <output_path>   : Optional. The path where the volume.toml file will be created.
                             If omitted, 'volume.toml' will be created in the current directory.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		outputFilePath := templateOutputArg

		if outputFilePath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting current working directory: %v\n", err)
				os.Exit(1)
			}
			outputFilePath = filepath.Join(cwd, "volume.toml")
			fmt.Printf("No output path specified. Defaulting to '%s'.\n", outputFilePath)
		}

		var libraryUUID string
		if templateFromArg != "" {
			resolvedUUID, err := persist.ResolveCollectionID(templateFromArg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not resolve '%s': %v. Using it directly as UUID.\n", templateFromArg, err)
				libraryUUID = templateFromArg
			} else {
				libraryUUID = resolvedUUID
			}
		}

		templateData := GenerateVolumeTemplate(libraryUUID)

		err := persist.SaveTomlConfig(templateData, outputFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating volume template: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created volume template at: %s\n", outputFilePath)
	},
}

func init() {
	templateCmd.AddCommand(templateNewCmd)

	templateNewCmd.Flags().StringVarP(&templateFromArg, "from", "f", "", "Optional: Specify UUID or shortname for the 'library' section. If a shortname is given, it will be resolved to its UUID from ~/.rsdish.")
	templateNewCmd.Flags().StringVarP(&templateOutputArg, "output", "o", "", "Optional: Path where the volume.toml file will be created. Defaults to 'volume.toml' in the current directory.")
}
