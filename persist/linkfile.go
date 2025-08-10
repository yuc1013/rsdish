package persist

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// LinkAll enumerates files in a source directory and creates links in a destination directory.
// It checks for existing files at the destination, but ignores symbolic links and "cheat files."
// The type of link created is determined by the `linkCreate` parameter.
func LinkAll(srcPath string, dstPath string, linkCreate string) error {
	if linkCreate == "none" {
		return nil // Do nothing if the linking mode is 'none'
	}

	// Make sure the destination directory exists
	if err := os.MkdirAll(dstPath, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory '%s': %w", dstPath, err)
	}

	// Walk through all files in the source directory
	err := filepath.WalkDir(srcPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Process only regular files
		if d.Type().IsRegular() {
			relPath, err := filepath.Rel(srcPath, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path for '%s': %w", path, err)
			}
			dstFilePath := filepath.Join(dstPath, relPath)

			// Check if a file with the same name already exists at the destination
			if _, err := os.Stat(dstFilePath); err == nil {
				// The file exists, but we need to check if it's a real file or a link/cheatfile.
				fileInfo, err := os.Lstat(dstFilePath)
				if err != nil {
					return fmt.Errorf("failed to get info for destination file '%s': %w", dstFilePath, err)
				}

				// If it's a symlink, treat it as not existing
				if fileInfo.Mode()&os.ModeSymlink != 0 {
					return createLink(path, dstFilePath, linkCreate)
				}

				// Read a small part of the file to check for "cheatfile" content
				content, readErr := os.ReadFile(dstFilePath)
				if readErr == nil && strings.TrimSpace(string(content)) == "cheatfile" {
					return createLink(path, dstFilePath, linkCreate)
				}

				// If it's a real file, we consider it a duplicate and do nothing
				if len(string(content)) < 20 {
					fmt.Printf("File '%s' already exists at destination, skipping. content: %s\n", dstFilePath, string(content))
				} else {
					fmt.Printf("File '%s' already exists at destination, skipping.\n", dstFilePath)
				}
				return nil
			}

			// If the file doesn't exist, create the link
			return createLink(path, dstFilePath, linkCreate)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk source directory '%s': %w", srcPath, err)
	}
	return nil
}

// createLink is a helper to build a symlink or cheatfile based on the mode.
func createLink(src string, dst string, mode string) error {
	// Ensure the parent directory for the link exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for link '%s': %w", dst, err)
	}

	switch mode {
	case "symlink":
		// Remove existing entry to prevent an error
		os.Remove(dst)
		if err := os.Symlink(src, dst); err != nil {
			return fmt.Errorf("failed to create symlink from '%s' to '%s': %w", src, dst, err)
		}
		fmt.Printf("Created symlink: %s -> %s\n", dst, src)
	case "cheatfile":
		// Remove existing entry
		os.Remove(dst)
		if err := os.WriteFile(dst, []byte("cheatfile"), 0644); err != nil {
			return fmt.Errorf("failed to create cheatfile at '%s': %w", dst, err)
		}
		fmt.Printf("Created cheatfile: %s\n", dst)
	default:
		// Should be unreachable due to the initial check in LinkAll
		return fmt.Errorf("invalid link creation mode: %s", mode)
	}

	return nil
}
