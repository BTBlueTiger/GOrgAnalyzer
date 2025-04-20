package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("🔍 Analyzing Git repo...")

	cmd := exec.Command("git", "log", "--pretty=%an")
	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(output), "\n")
	commitCounts := make(map[string]int)

	for _, line := range lines {
		if line != "" {
			commitCounts[line]++
		}
	}

	fmt.Println("📊 Commits by author:")
	for author, count := range commitCounts {
		fmt.Printf("👤 %s: %d\n", author, count)
	}
}