package main

import (
	"bufio"
	"fmt"
	
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var extToLang = map[string]string{
	".go":    "Go",
	".ts":    "TypeScript",
	".cs":    "C#",
	".py":    "Python",
	".java":  "Java",
	".js":    "JavaScript",
	".cpp":   "C++",
	".c":     "C",
	".rb":    "Ruby",
	".php":   "PHP",
	".html":  "HTML",
	".css":   "CSS",
	".rs":    "Rust",
	".swift": "Swift",
	".kt":    "Kotlin",
	".sh":    "Shell",
	".xml":   "XML",
	".yaml":  "YAML",
	".yml":   "YAML",}

// shouldIgnorePath checks if a given path matches any patterns in the .gitignore file for the current directory.
func shouldIgnorePath(repoPath, path string) (bool, error) {
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // No .gitignore file, so nothing to ignore
		}
		return false, fmt.Errorf("opening .gitignore: %w", err)
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		patterns = append(patterns, line)
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("reading .gitignore: %w", err)
	}

	relativePath, err := filepath.Rel(repoPath, path)
	if err != nil {
		return false, fmt.Errorf("getting relative path: %w", err)
	}

	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, relativePath)
		if err != nil {
			return false, fmt.Errorf("matching pattern: %w", err)
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

// analyzeLanguages calculates the percentage of each programming language in a directory based on the number of bytes in the files.
func analyzeLanguages(repoPath string) (map[string]int, int, error) {
	langByteCounts := make(map[string]int)
	totalBytes := 0

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Warning: Skipping invalid path %s: %v", path, err)
			return nil
		}
		if info.IsDir() {
			return nil
		}

		ignore, err := shouldIgnorePath(repoPath, path)
		if err != nil {
			log.Printf("Error checking .gitignore for %s: %v", path, err)
			return nil
		}
		if ignore {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if lang, exists := extToLang[ext]; exists {
			file, err := os.Open(path)
			if err != nil {
				log.Printf("Warning: Unable to open file %s: %v", path, err)
				return nil
			}
			defer file.Close()

			stat, err := file.Stat()
			if err != nil {
				log.Printf("Warning: Unable to get file info for %s: %v", path, err)
				return nil
			}

			byteCount := int(stat.Size())
			langByteCounts[lang] += byteCount
			totalBytes += byteCount
		}
		return nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("walking file path: %w", err)
	}

	return langByteCounts, totalBytes, nil
}

// analyzeCommitsByAuthor counts commits by author in a Git repository.
func analyzeCommitsByAuthor(repoPath string) (map[string]int, error) {
	cmd := exec.Command("git", "-C", repoPath, "log", "--pretty=%an")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running git log: %w", err)
	}

	counts := make(map[string]int)
	for _, author := range strings.Split(string(output), "\n") {
		if author != "" {
			counts[author]++
		}
	}
	return counts, nil
}

// processGitRepo analyzes a single Git repository for commits and languages.
func processGitRepo(repoPath string, totalLangCounts map[string]int, totalFilesAnalyzed *int) error {
	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("*********************************************\n")
	fmt.Printf("🔍 Analyzing Git repository: %s\n", repoPath)

	// Analyze commits by author
	commitCounts, err := analyzeCommitsByAuthor(repoPath)
	if err != nil {
		log.Printf("Error analyzing commits in %s: %v", repoPath, err)
		return nil
	}

	fmt.Println("📊 Commits by author:")
	for author, count := range commitCounts {
		fmt.Printf("👤 %s: %d\n", author, count)
	}

	// Analyze programming languages
	repoLangCounts, repoTotalBytes, err := analyzeLanguages(repoPath)
	if err != nil {
		log.Printf("Error analyzing languages in %s: %v", repoPath, err)
		return nil
	}

	// Print language statistics
	fmt.Println("📊 Language statistics:")
	for lang, byteCount := range repoLangCounts {
		percentage := (float64(byteCount) / float64(repoTotalBytes)) * 100
		fmt.Printf("📝 %s: %d bytes (%.2f%%)\n", lang, byteCount, percentage)
	}

	// Update cumulative statistics
	for lang, count := range repoLangCounts {
		totalLangCounts[lang] += count
	}
	*totalFilesAnalyzed += repoTotalBytes

	return nil
}
