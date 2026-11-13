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
