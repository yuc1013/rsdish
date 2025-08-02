package persist

import (
	"fmt"
	"strings"
)

// RcloneOptions holds the necessary information to construct an Rclone command.
// RcloneArguments is expected to be a string of space-separated Rclone flags (e.g., "--dry-run --progress").
type RcloneOptions struct {
	Src             string // Source path for the rclone command
	Dst             string // Destination path for the rclone command
	RcloneArguments string // Raw string of rclone flags/arguments for the command
}

// BuildRcloneCommands takes a slice of RcloneOptions and returns a slice of complete
// Rclone command strings, suitable for execution. It calls BuildRcloneCommand for each option.
func BuildRcloneCommands(cmds []RcloneOptions) []string {
	rcloneCmds := make([]string, len(cmds))
	for i, opt := range cmds {
		rcloneCmds[i] = BuildRcloneCommand(opt)
	}
	return rcloneCmds
}

// BuildRcloneCommand constructs a single Rclone command string using 'rclone copy'.
// Rclone is cross-platform and generally prefers forward slashes for paths.
// Assumes 'rclone' executable is in the system's PATH.
func BuildRcloneCommand(options RcloneOptions) string {
	// Quoting paths to handle spaces and ensure robustness if the command string
	// is directly used in a shell.
	src := fmt.Sprintf(`"%s"`, options.Src)
	dst := fmt.Sprintf(`"%s"`, options.Dst)

	// The RcloneArguments string is trimmed for whitespace and then appended directly.
	rcloneArgs := strings.TrimSpace(options.RcloneArguments)

	// *** Key Change: Using 'rclone copy' instead of 'rclone sync' ***
	// The base command will always be "rclone copy" followed by source, destination,
	// and any additional arguments.
	cmd := fmt.Sprintf("rclone copy %s %s", src, dst)

	// Append RcloneArguments only if they are not empty, to avoid trailing spaces.
	if rcloneArgs != "" {
		cmd = fmt.Sprintf("%s %s", cmd, rcloneArgs)
	}

	return cmd
}
