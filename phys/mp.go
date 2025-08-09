package phys

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// getWindowsMountPoints enumerates drive letters from A: to Z: directly.
func getWindowsMountPoints() ([]string, error) {
	var mountPoints []string
	for c := 'A'; c <= 'Z'; c++ {
		drive := string(c) + ":\\"
		// 判断盘符是否存在（即路径存在）
		if _, err := os.Stat(drive); err == nil {
			mountPoints = append(mountPoints, filepath.Clean(drive)+"\\")
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
