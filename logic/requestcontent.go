package logic

import (
	"os"
	"yact/config"
	"yact/projectpath"
)

type RequestContent struct {
	Prompt string
	Files  []projectpath.Path
	Domain projectpath.Path
	Index  projectpath.Path
}

const (
	domainFileName = "domain.txt"
	indexFileName  = "index.txt"
	bufferFileName = "buffer.txt"
)

func IndexFilePath() projectpath.Path {
	return projectpath.Join(config.GetProjectTempDir(), indexFileName)
}

func BufferFilePath() projectpath.Path {
	return projectpath.Join(config.GetProjectTempDir(), bufferFileName)
}

func RemoveIndexFile() error {
	if err := os.Remove(string(IndexFilePath())); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
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
	for _, path := range []projectpath.Path{r.Domain, r.Index} {
		if path != "" {
			paths = append(paths, path)
		}
	}

	return paths
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

	return RequestContent{
		Prompt: "",
		Files:  []projectpath.Path{},
		Domain: domain,
		Index:  index,
	}, nil
}

func messageFilePaths() []projectpath.Path {
	return []projectpath.Path{
		IndexFilePath(),
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
