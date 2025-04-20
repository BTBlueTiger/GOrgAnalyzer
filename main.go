package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kamalte/GOrgAnalyzer/analyze"
)

// GitHub language colors
var githubLangColors = map[string]string{
	"Go":          "#00ADD8",
	"TypeScript":  "#3178C6",
	"C#":          "#5C2D91",
	"Python":      "#3776AB",
	"Java":        "#b07219",
	"JavaScript":  "#f1e05a",
	"C++":         "#f34b7d",
	"C":           "#555555",
	"Ruby":        "#701516",
	"PHP":         "#4F5D95",
	"HTML":        "#E34C26",
	"CSS":         "#264DE4",
	"Rust":        "#dea584",
	"Swift":       "#ffac45",
	"Kotlin":      "#A97BFF",
	"Shell":       "#89E051",
	"XML":         "#0060ac",
	"YAML":        "#8A2BE2",
}

// generateProgressBarSVG creates an SVG graphic with a progress bar for language usage.
func generateProgressBarSVG(langByteCounts map[string]int, totalBytes int, outputPath string) error {
	const svgHeader = `<svg xmlns="http://www.w3.org/2000/svg" width="600" height="70" viewBox="0 0 600 70" style="background-color:#000000; font-family:Arial, sans-serif; border:2px solid #ffffff; border-radius:5px;">`
	const svgFooter = `</svg>`

	var svgContent strings.Builder
	svgContent.WriteString(svgHeader)

	// Add title
	

	// Sort the languages by size in descending order
	type langData struct {
		lang      string
		byteCount int
	}
	var sortedLangs []langData
	for lang, byteCount := range langByteCounts {
		sortedLangs = append(sortedLangs, langData{lang, byteCount})
	}
	sort.Slice(sortedLangs, func(i, j int) bool {
		return sortedLangs[i].byteCount > sortedLangs[j].byteCount
	})

	// Variables for progress bar
	x, y, barHeight := 50.0, 20.0, 10.0
	totalWidth := 500.0
	currentX := x

	// Generate progress bar segments
	for i, data := range sortedLangs {
		percentage := float64(data.byteCount) / float64(totalBytes)
		barWidth := totalWidth * percentage

		// Use GitHub color for the language or fallback to a random color
		color, exists := githubLangColors[data.lang]
		if !exists {
			color = fmt.Sprintf("#%06x", rand.Intn(0xFFFFFF))
		}

		// Determine rounded corners for the first and last segments
		rxLeft, ryRight := 0.0, 0.0
		if i == 0 {
			rxLeft = 10.0 // Rounded corners for the left side of the first segment
		}
		if i == len(sortedLangs)-1 {
			ryRight = 10.0 // Rounded corners for the right side of the last segment
		}

		// Add the rectangle for the segment
		svgContent.WriteString(fmt.Sprintf(
			`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="%.2f" ry="%.2f" fill="%s" />`,
			currentX, y, barWidth, barHeight, rxLeft, ryRight, color,
		))

		currentX += barWidth
	}

	svgContent.WriteString(svgFooter)

	// Write the SVG content to the output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating SVG file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(svgContent.String())
	if err != nil {
		return fmt.Errorf("writing SVG content: %w", err)
	}

	return nil
}

// main orchestrates the analysis of Git repositories.
func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a path to the base directory to analyze.")
	}

	basePath, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatalf("Error resolving absolute path: %v", err)
	}

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		log.Fatalf("Error: The provided path '%s' does not exist.", basePath)
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		log.Fatalf("Error reading base directory: %v", err)
	}

	totalLangCounts := make(map[string]int)
	totalBytesAnalyzed := 0

	for _, entry := range entries {
		if entry.IsDir() {
			repoPath := filepath.Join(basePath, entry.Name())
			if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
				continue // Skip if not a Git repository
			}

			if err := analyze.ProcessGitRepo(repoPath, totalLangCounts, &totalBytesAnalyzed); err != nil {
				log.Printf("Error processing repository %s: %v", repoPath, err)
			}
		}
	}

	// Output final summary
	if totalBytesAnalyzed > 0 {
		fmt.Println("\n📊 Final Summary of Programming Languages Across All Repositories:")
		for lang, count := range totalLangCounts {
			percentage := float64(count) / float64(totalBytesAnalyzed) * 100
			fmt.Printf("%s: %.2f%% (%d bytes)\n", lang, percentage, count)
		}

		// Generate cumulative progress bar SVG graphic
		progressBarOutputPath := filepath.Join(basePath, "./cumulative_language_progress_bar.svg")
		err = generateProgressBarSVG(totalLangCounts, totalBytesAnalyzed, progressBarOutputPath)
		if err != nil {
			log.Printf("Error generating progress bar SVG: %v", err)
		} else {
			fmt.Printf("📈 Progress bar SVG graphic generated at: %s\n", progressBarOutputPath)
		}
	} else {
		fmt.Println("\nNo programming files were analyzed across the repositories.")
	}
}
