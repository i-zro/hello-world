package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"regexp"
	"sort"
)

func extractMarkdownTables(mdText string) [][]string {
	tablePattern := regexp.MustCompile(`(?m)\|(.+?)\|\n\|[-| :]+?\|\n((?:\|.*\|\n)*)`)
	tables := tablePattern.FindAllStringSubmatch(mdText, -1)

	var result [][]string
	for _, table := range tables {
		rows := regexp.MustCompile(`(?m)\|(.+)\|`).FindAllStringSubmatch(table[0], -1)
		var tableData []string
		for _, row := range rows {
			if strings.Contains(row[1], "[습관]") {
				tableData = append(tableData, row[1])
			}
		}
		result = append(result, tableData)
	}

	return result
}

func updateOrCreateHabitFile(path string, fileName string, habitData map[string][]string) {
	// Open the existing file or create a new one.
	file, err := os.OpenFile(path + "habit-" + fileName + ".md", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	// Write the habit data to the file.
	var habitNames []string
	for habitName := range habitData {
		habitNames = append(habitNames, habitName)
	}
	sort.Strings(habitNames)
	for _, habitName := range habitNames {
		file.WriteString(fmt.Sprintf("\n# %s\n| 날짜 | 완료여부 |\n|--|--|\n", habitName))
		for _, data := range habitData[habitName] {
			file.WriteString(data)
		}
	}
}

func main() {
	// github actions와 로컬 환경 경로 분리
	var rootDir string
	var viewDir string
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		rootDir = os.Getenv("GITHUB_WORKSPACE") + "/archieve/"
		viewDir = os.Getenv("GITHUB_WORKSPACE") + "/view/"
	} else {
		rootDir = "./archieve/"
		viewDir = "./view/"
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

			// Extract tables.
			tables := extractMarkdownTables(mdText)

			if len(tables) > 0 {
				fileName := filepath.Base(path)
				fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
				fileYearMonth := fileName[:4]
				
				habitData := make(map[string][]string)
				
				for _, table := range tables {
					if len(table) > 0 {
						for _, row := range table {
							columns := strings.Split(row, "|")
							habitName := strings.TrimSpace(columns[2])
							habitStatus := strings.TrimSpace(columns[3])
							habitData[habitName] = append(habitData[habitName], fmt.Sprintf("| %s | %s |\n", fileName, habitStatus))
						}
					}
				}

				updateOrCreateHabitFile(viewDir, fileYearMonth, habitData)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
	}
}
