package logic

import (
	"os"
	"path/filepath"
	"yact/config"
)

type RequestContent struct {
	Prompt string
	Files  []string
	Domain string
	Index  string
	Output string
}

const (
	domainFileName = "domain.txt"
	indexFileName  = "index.txt"
	outputFileName = "output.txt"
)

var messageFileNames = []string{domainFileName, indexFileName, outputFileName}

func yactDir() string {
	return config.GetProjectYactDir()
}

func IndexFilePath() string {
	return filepath.Join(yactDir(), indexFileName)
}

func RemoveIndexFile() error {
	if err := os.Remove(IndexFilePath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeFixedFile(name string, content string) (string, error) {
	if err := os.MkdirAll(yactDir(), 0755); err != nil {
		return "", err
	}

	path := filepath.Join(yactDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}

	return path, nil
}

func existingFilePath(name string) (string, error) {
	path := filepath.Join(yactDir(), name)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	return path, nil
}

func (r RequestContent) MessagePaths() []string {
	paths := []string{}
	for _, path := range []string{r.Domain, r.Index, r.Output} {
		if path != "" {
			paths = append(paths, path)
		}
	}

	return paths
}

func AddMessage(content string) (string, error) {
	return writeFixedFile(domainFileName, content)
}

func AddOutput(content string) (string, error) {
	return writeFixedFile(outputFileName, content)
}

func LoadRequestContent() (RequestContent, error) {
	domain, err := existingFilePath(domainFileName)
	if err != nil {
		return RequestContent{}, err
	}

	index, err := existingFilePath(indexFileName)
	if err != nil {
		return RequestContent{}, err
	}

	output, err := existingFilePath(outputFileName)
	if err != nil {
		return RequestContent{}, err
	}

	return RequestContent{
		Prompt: "",
		Files:  []string{},
		Domain: domain,
		Index:  index,
		Output: output,
	}, nil
}

func RemoveMessages() error {
	for _, name := range messageFileNames {
		if err := os.Remove(filepath.Join(yactDir(), name)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}
