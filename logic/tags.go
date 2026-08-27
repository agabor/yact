package logic

import (
	"encoding/csv"
	"os"
	"sort"
	"strings"

	"yact/config"
)

func LoadTags() (map[string][]string, error) {
	tagsPath := config.GetProjectTagsPath()

	data, err := os.ReadFile(tagsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string][]string), nil
		}
		return nil, err
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	tags := make(map[string][]string)
	for _, record := range records {
		if len(record) == 0 {
			continue
		}
		filePath := record[0]
		fileTags := record[1:]
		tags[filePath] = fileTags
	}

	return tags, nil
}

func SaveTags(tags map[string][]string) error {
	yactDir := config.GetProjectYactDir()
	if err := os.MkdirAll(yactDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(config.GetProjectTagsPath())
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var filePaths []string
	for filePath := range tags {
		filePaths = append(filePaths, filePath)
	}
	sort.Strings(filePaths)

	for _, filePath := range filePaths {
		record := append([]string{filePath}, tags[filePath]...)
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func GetFilesByTag(tagName string) ([]string, error) {
	tags, err := LoadTags()
	if err != nil {
		return nil, err
	}

	var files []string
	for filePath, fileTags := range tags {
		for _, tag := range fileTags {
			if tag == tagName {
				files = append(files, filePath)
				break
			}
		}
	}

	sort.Strings(files)

	return files, nil
}