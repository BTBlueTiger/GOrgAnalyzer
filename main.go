package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kamalte/GOrgAnalyzer/analyze"
)

// LoadGitLangColors loads the language colors from a JSON file.
func LoadGitLangColors(filePath string) (map[string]string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var langColors map[string]string
	if err := json.Unmarshal(data, &langColors); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return langColors, nil
}

func generateProgressBarSVG(langByteCounts map[string]int, totalBytes int, outputPath string, githubLangColors map[string]string) error {
	const svgHeader = `<svg xmlns="http://www.w3.org/2000/svg" width="800" height="20" style="font-family:Arial, sans-serif;">`
	const svgFooter = `</svg>`

	var svgContent strings.Builder
	svgContent.WriteString(svgHeader)

	// Create a clipPath for the rounded corners
	svgContent.WriteString(`
		<defs>
			<clipPath id="roundedClip">
				<rect x="0" y="0" width="800" height="20" rx="10" ry="10"/>
			</clipPath>
		</defs>
	`)

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
	x, barHeight := 0.0, 20.0 // Increased bar height for better visibility
	totalWidth := 800.0
	currentX := x

	// Start the progress bar group, applying the clipPath
	svgContent.WriteString(`<g clip-path="url(#roundedClip)">`)

	// Generate progress bar segments
	for _, data := range sortedLangs {
		percentage := float64(data.byteCount) / float64(totalBytes)
		barWidth := totalWidth * percentage

		// Use GitHub color for the language or fallback to a random color
		color, exists := githubLangColors[data.lang]
		if !exists {
			color = fmt.Sprintf("#%06x", rand.Intn(0xFFFFFF))
		}

		// Add the rectangle for the segment
		svgContent.WriteString(fmt.Sprintf(
			`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s" />`,
			currentX, 0.0, barWidth, barHeight, color,
		))
		currentX += barWidth
	}

	// End the group element that applies the clipPath
	svgContent.WriteString(`</g>`)

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

	// Load GitHub language colors from JSON file
	langColorsPath := "./git_lang_colors.json"
	githubLangColors, err := LoadGitLangColors(langColorsPath)
	if err != nil {
		log.Fatalf("Error loading language colors: %v", err)
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
		progressBarOutputPath := "./cumulative_language_progress_bar.svg"
		err = generateProgressBarSVG(totalLangCounts, totalBytesAnalyzed, progressBarOutputPath, githubLangColors)
		if err != nil {
			log.Printf("Error generating progress bar SVG: %v", err)
		} else {
			fmt.Printf("📈 Progress bar SVG graphic generated at: %s\n", progressBarOutputPath)
		}
	} else {
		fmt.Println("\nNo programming files were analyzed across the repositories.")
	}
}
