package phys

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// GetMountPoints returns a list of all currently mounted paths for the current OS.
func GetMountPoints() ([]string, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		return getUnixMountPoints()
	case "windows":
		return getWindowsMountPoints()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// getUnixMountPoints fetches mount points on macOS and Linux using 'df -h'.
func getUnixMountPoints() ([]string, error) {
	cmd := exec.Command("df", "-h") // -h for human-readable, not strictly necessary for paths
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run 'df -h': %w", err)
	}

	lines := strings.Split(out.String(), "\n")
	var mountPoints []string
	// Skip header and empty lines
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}
		fields := strings.Fields(line)
		if len(fields) < 6 { // Expect at least Filesystem, Size, Used, Avail, Use%, Mounted on
			continue
		}
		// The mount point is usually the last field
		mountPoint := fields[len(fields)-1]

		// Basic filter for potential unwanted lines (e.g., special file systems that don't need listing)
		if strings.HasPrefix(mountPoint, "/dev") || strings.HasPrefix(mountPoint, "tmpfs") ||
			strings.HasPrefix(mountPoint, "overlay") || strings.HasPrefix(mountPoint, "none") {
			// You might want to fine-tune this filter based on what you consider "relevant" mount points
			continue
		}

		mountPoints = append(mountPoints, mountPoint)
	}

	// Ensure root is always included if not caught by df output for some reason (unlikely but safe)
	if !contains(mountPoints, "/") && (runtime.GOOS == "linux" || runtime.GOOS == "darwin") {
		mountPoints = append(mountPoints, "/")
	}

	return mountPoints, nil
}

// getWindowsMountPoints fetches mount points (drive letters) on Windows using 'wmic'.
func getWindowsMountPoints() ([]string, error) {
	// Using wmic to get logical disk info. /format:csv makes it easier to parse.
	cmd := exec.Command("wmic", "logicaldisk", "get", "Caption,Freespace,Size", "/format:csv")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run 'wmic logicaldisk get Caption,Freespace,Size /format:csv': %w", err)
	}

	lines := strings.Split(out.String(), "\n")
	var mountPoints []string

	// Skip first line (Node) and second line (header)
	for i, line := range lines {
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}
		// Example line: ",C:,26012720128,142998675456" (empty first field, Caption, Freespace, Size)
		// We're interested in the second field (Caption) which is the drive letter.
		fields := strings.Split(line, ",")
		if len(fields) >= 2 {
			driveLetter := strings.TrimSpace(fields[1])
			if driveLetter != "" && strings.HasSuffix(driveLetter, ":") { // Ensure it's like "C:"
				mountPoints = append(mountPoints, driveLetter+"\\") // Append backslash for Windows path
			}
		}
	}
	return mountPoints, nil
}

// contains is a helper to check if a string is in a slice.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
