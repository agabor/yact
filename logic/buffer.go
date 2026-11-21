package logic

import (
	"encoding/json"
	"os"
	"yact/config"
)

func ensureBufferDir() error {
	return os.MkdirAll(string(config.GetProjectTempDir()), 0755)
}

func PutBuffer(session Session) error {
	if err := ensureBufferDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.Create(string(BufferFilePath()))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func ReadBuffer() (Session, error) {
	data, err := os.ReadFile(string(BufferFilePath()))
	if err != nil {
		if os.IsNotExist(err) {
			return Session{}, nil
		}
		return Session{}, err
	}
	return parseBufferContent(data), nil
}

// Buffers written before sessions existed hold plain response text; they have no linked files, so only feedback.md can be written from them.
func parseBufferContent(data []byte) Session {
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return Session{Response: string(data)}
	}
	return session
}
