package logic

import (
	"os"
	"yact/config"
	"yact/projectpath"
)

type RequestContent struct {
	Prompt string
	Files  []CodeFile
	Domain CodeFile
	Index  CodeFile
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

func LoadRequestContent() (*RequestContent, error) {
	domain, err := existingFilePath(projectpath.Join(config.GetProjectYactDir(), domainFileName))
	if err != nil {
		return &RequestContent{}, err
	}

	index, err := existingFilePath(IndexFilePath())
	if err != nil {
		return &RequestContent{}, err
	}

	domainFile, err := ReadAsFile(domain)
	if err != nil {
		return &RequestContent{}, err
	}
	indexFile, err := ReadAsFile(index)
	if err != nil {
		return &RequestContent{}, err
	}

	return &RequestContent{
		Prompt: "",
		Files:  []CodeFile{},
		Domain: domainFile,
		Index:  indexFile,
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
