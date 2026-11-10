package logic

import (
	"testing"
)

func TestIndexFilePath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "index file is in temp dir",
			want: ".yact/temp/index.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(IndexFilePath()); got != tt.want {
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
			want: ".yact/temp/buffer.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(BufferFilePath()); got != tt.want {
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
			want: ".yact/temp/output.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(outputFilePath()); got != tt.want {
				t.Errorf("outputFilePath() = %q, want %q", got, tt.want)
			}
		})
	}
}
