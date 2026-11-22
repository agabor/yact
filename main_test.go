package main

import (
	"errors"
	"strings"
	"testing"

	"yact/api"
	"yact/config"
)

func TestRequireArgCount(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		counts  []int
		wantErr bool
	}{
		{name: "matching count", args: []string{"a"}, counts: []int{1}, wantErr: false},
		{name: "non-matching count", args: []string{}, counts: []int{1}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireArgCount("cmd", tt.args, tt.counts...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("requireArgCount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequireNoArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "zero arguments", args: []string{}, wantErr: false},
		{name: "one argument", args: []string{"x"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireNoArgs("cmd", tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("requireNoArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPromptNotFoundError(t *testing.T) {
	err := promptNotFoundError(config.PromptName("review"), errors.New("network down"))

	if !strings.Contains(err.Error(), "review") {
		t.Errorf("error %q does not name the prompt", err.Error())
	}
	if !strings.Contains(err.Error(), "network down") {
		t.Errorf("error %q does not contain the download error", err.Error())
	}
}

func TestRun(t *testing.T) {
	failingFactory := func(cfg *config.Config) api.Client {
		t.Fatal("client factory must not be called")
		return nil
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "help", args: []string{"help"}, wantErr: false},
		{name: "version flag", args: []string{"--version"}, wantErr: false},
		{name: "no command", args: []string{}, wantErr: true},
		{name: "command takes no arguments", args: []string{"act", "foo"}, wantErr: true},
		{name: "keyword without argument", args: []string{"keyword"}, wantErr: true},
		{name: "prompt without argument", args: []string{"prompt"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.args, failingFactory)
			if (err != nil) != tt.wantErr {
				t.Fatalf("run(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}
