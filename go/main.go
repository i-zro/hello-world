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

func writeHabitFile(path string, data map[string][]string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for habit, rows := range data {
		_, err = file.WriteString(fmt.Sprintf("\n# %s\n| 날짜 | 완료여부 |\n|--|--|\n", habit))
		if err != nil {
			return err
		}

		for _, row := range rows {
			_, err = file.WriteString(fmt.Sprintf("%s\n", row))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	// github actions와 로컬 환경 경로 분리
	var rootDir, habitFile string
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		rootDir = os.Getenv("GITHUB_WORKSPACE") + "/archieve/"
		habitFile = os.Getenv("GITHUB_WORKSPACE") + "/view/habit.md"
	} else {
		rootDir = "./archieve/"
		habitFile = "./view/habit.md"
	}

	result := make(map[string][]string)
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

			// Extract the filename without extension.
			_, filename := filepath.Split(path)
			filename = strings.TrimSuffix(filename, filepath.Ext(filename))

			// Extract tables.
			tables := extractMarkdownTables(mdText, filename)

			// Merge the results.
			for k, v := range tables {
				result[k] = append(result[k], v...)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Write the results to the habit file.
	err = writeHabitFile(habitFile, result)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
