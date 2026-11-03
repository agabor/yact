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
	// Set by the caller for a single call and never persisted, so the skills of one
	// call cannot leak into the next one.
	Skills []Skill
}

type Skill struct {
	Name    string
	Content string
}

const (
	domainFileName = "domain.txt"
	indexFileName  = "index.txt"
	outputFileName = "output.txt"
	bufferFileName = "buffer.txt"
)

func yactDir() string {
	return config.GetProjectYactDir()
}

func tempDir() string {
	return filepath.Join(yactDir(), "temp")
}

func IndexFilePath() string {
	return filepath.Join(tempDir(), indexFileName)
}

func BufferFilePath() string {
	return filepath.Join(tempDir(), bufferFileName)
}

func outputFilePath() string {
	return filepath.Join(tempDir(), outputFileName)
}

func RemoveIndexFile() error {
	if err := os.Remove(IndexFilePath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeFixedFile(path string, content string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}

	return path, nil
}

func existingFilePath(path string) (string, error) {
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
	return writeFixedFile(filepath.Join(yactDir(), domainFileName), content)
}

func AddOutput(content string) (string, error) {
	return writeFixedFile(outputFilePath(), content)
}

func LoadRequestContent() (RequestContent, error) {
	domain, err := existingFilePath(filepath.Join(yactDir(), domainFileName))
	if err != nil {
		return RequestContent{}, err
	}

	index, err := existingFilePath(IndexFilePath())
	if err != nil {
		return RequestContent{}, err
	}

	output, err := existingFilePath(outputFilePath())
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

func messageFilePaths() []string {
	return []string{
		IndexFilePath(),
		outputFilePath(),
	}
}

func RemoveMessages() error {
	for _, path := range messageFilePaths() {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}
