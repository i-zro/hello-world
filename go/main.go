package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"regexp"
	"unicode"
)

func cleanString(str string) string {
    return strings.Map(func(r rune) rune {
        if unicode.IsPrint(r) {
            return r
        }
        return -1
    }, str)
}

func extractMarkdownTables(mdText string, filename string) map[string][]string {
	tablePattern := regexp.MustCompile(`(?m)\|(.+?)\|\n\|[-| :]+?\|\n((?:\|.*\|\n)*)`)
	tables := tablePattern.FindAllStringSubmatch(mdText, -1)

	result := make(map[string][]string)
	for _, table := range tables {
		rows := regexp.MustCompile(`(?m)\|(.+)\|`).FindAllStringSubmatch(table[0], -1)
		for _, row := range rows {
			if strings.Contains(row[1], "[습관]") {
				columns := strings.Split(row[1], "|")
				if len(columns) > 3 {
					habit := strings.TrimSpace(columns[2])
					status := strings.TrimSpace(columns[3])
					result[habit] = append(result[habit], fmt.Sprintf("| %s | %s |", filename, cleanString(status)))
				}
			}
		}
	}

	return result
}

func main() {
	// github actions와 로컬 환경 경로 분리
	var rootDir string
	var outputDir string
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		rootDir = os.Getenv("GITHUB_WORKSPACE") + "/archive/"
		outputDir = os.Getenv("GITHUB_WORKSPACE") + "/view/"
	} else {
		rootDir = "./archive/"
		outputDir = "./view/"
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Current file is a directory, skip it.
		if info.IsDir() {
			return nil
		}

		// Check if the file is a Markdown file.
		if strings.HasSuffix(path, ".md") {
			// Read the Markdown file.
			mdBytes, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			mdText := string(mdBytes)
			filename := strings.TrimSuffix(info.Name(), ".md")

			// Extract tables.
			tables := extractMarkdownTables(mdText, filename)

			// Open output file, create if not exists.
			file, err := os.OpenFile(outputDir+"habit.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer file.Close()

			// Write to file.
			for habit, rows := range tables {
				if _, err := file.WriteString("# " + habit + "\n"); err != nil {
					return err
				}
				if _, err := file.WriteString("| 날짜 | 완료여부 |\n|--|--|\n"); err != nil {
					return err
				}
				for _, row := range rows {
					if _, err := file.WriteString(row + "\n"); err != nil {
						return err
					}
				}
				if _, err := file.WriteString("\n"); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
	}
}
