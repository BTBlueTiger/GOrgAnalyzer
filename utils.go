package main

import (
	"fmt"
	"os"
)

// listDirs prints all directories inside the given path.
func listDirs(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	fmt.Println("📂 Directories inside", path)
	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Println("📁", entry.Name())
		}
	}
	return nil
}
