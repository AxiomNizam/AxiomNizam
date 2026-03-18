//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"regexp"
	"time"
)

var svgIcons = []string{
	`<svg viewBox="0 0 24 24"><rect x="18" y="3" width="4" height="18" rx="1"/><rect x="10" y="8" width="4" height="13" rx="1"/><rect x="2" y="13" width="4" height="8" rx="1"/></svg>`,
	`<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/><path d="M2 12h20"/></svg>`,
	`<svg viewBox="0 0 24 24"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/></svg>`,
	`<svg viewBox="0 0 24 24"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>`,
	`<svg viewBox="0 0 24 24"><path d="M2 12a10 10 0 0 1 10-10"/><path d="M2 16a6 6 0 0 1 6-6"/><path d="M2 20a2 2 0 0 1 2-2"/><path d="M19 19m-3 0a3 3 0 1 0 6 0 3 3 0 1 0-6 0"/><path d="M16 16l-4-4"/></svg>`,
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Running from within the frontend folder
	files, err := filepath.Glob("templates/*.html")
	if err != nil {
		fmt.Println("Error finding files:", err)
		return
	}

	navRegex := regexp.MustCompile(`<span class="nav-icon">([^<]+)</span>`)
	tabRegex := regexp.MustCompile(`<span class="tab-icon">([^<]+)</span>`)

	var successCount = 0
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", file, err)
			continue
		}

		strContent := string(content)

		// Replace nav-icons
		strContent = navRegex.ReplaceAllStringFunc(strContent, func(match string) string {
			randomSVG := svgIcons[rand.Intn(len(svgIcons))]
			return fmt.Sprintf(`<span class="nav-icon">%s</span>`, randomSVG)
		})

		// Replace tab-icons
		strContent = tabRegex.ReplaceAllStringFunc(strContent, func(match string) string {
			randomSVG := svgIcons[rand.Intn(len(svgIcons))]
			return fmt.Sprintf(`<span class="tab-icon">%s</span>`, randomSVG)
		})

		// Write modified content back
		err = ioutil.WriteFile(file, []byte(strContent), 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", file, err)
		} else {
			successCount++
		}
	}
	fmt.Printf("Done! Processed %d templates.\n", successCount)
}
