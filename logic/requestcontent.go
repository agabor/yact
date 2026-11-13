package logic

import (
	"os"
	"path/filepath"
	"yact/config"
	"yact/projectpath"
)

type RequestContent struct {
	Prompt string
	Files  []projectpath.Path
	Domain projectpath.Path
	Index  projectpath.Path
	Output projectpath.Path
}

const (
	domainFileName = "domain.txt"
	indexFileName  = "index.txt"
	outputFileName = "output.txt"
	bufferFileName = "buffer.txt"
)

func IndexFilePath() projectpath.Path {
	return projectpath.Join(config.GetProjectTempDir(), indexFileName)
}

func BufferFilePath() projectpath.Path {
	return projectpath.Join(config.GetProjectTempDir(), bufferFileName)
}

func outputFilePath() projectpath.Path {
	return projectpath.Join(config.GetProjectTempDir(), outputFileName)
}

func RemoveIndexFile() error {
	if err := os.Remove(string(IndexFilePath())); err != nil && !os.IsNotExist(err) {
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

func existingFilePath(path projectpath.Path) (projectpath.Path, error) {
	if _, err := os.Stat(string(path)); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	return path, nil
}

func (r RequestContent) MessagePaths() []projectpath.Path {
	paths := []projectpath.Path{}
	for _, path := range []projectpath.Path{r.Domain, r.Index, r.Output} {
		if path != "" {
			paths = append(paths, path)
		}
	}

	return paths
}

func AddMessage(content string) (string, error) {
	return writeFixedFile(string(projectpath.Join(config.GetProjectYactDir(), domainFileName)), content)
}

func AddOutput(content string) (string, error) {
	return writeFixedFile(string(outputFilePath()), content)
}

func LoadRequestContent() (RequestContent, error) {
	domain, err := existingFilePath(projectpath.Join(config.GetProjectYactDir(), domainFileName))
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
		Files:  []projectpath.Path{},
		Domain: domain,
		Index:  index,
		Output: output,
	}, nil
}

func messageFilePaths() []projectpath.Path {
	return []projectpath.Path{
		IndexFilePath(),
		outputFilePath(),
	}
}

func RemoveMessages() error {
	for _, path := range messageFilePaths() {
		if err := os.Remove(string(path)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}
