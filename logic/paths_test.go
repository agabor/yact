package logic

import (
	"path/filepath"
	"testing"
)

func TestIndexFilePath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "index file is in temp dir",
			want: filepath.Join(".yact", "temp", "index.txt"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IndexFilePath(); got != tt.want {
				t.Errorf("IndexFilePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBufferFilePath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "buffer file is in temp dir",
			want: filepath.Join(".yact", "temp", "buffer.txt"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BufferFilePath(); got != tt.want {
				t.Errorf("BufferFilePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOutputFilePath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "output file is in temp dir",
			want: filepath.Join(".yact", "temp", "output.txt"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := outputFilePath(); got != tt.want {
				t.Errorf("outputFilePath() = %q, want %q", got, tt.want)
			}
		})
	}
}