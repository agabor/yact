package config

import (
	"github.com/spf13/pflag"
)

// CLIFlags holds the raw command line values until ToFlaglist parses them.
type CLIFlags struct {
	Help       bool
	Version    bool
	Model      string
	Output     string
	Strict     bool
	Index      string
	NoProgress bool
	Verbose    bool
	DryRun     bool
	Batch      bool
}

func RegisterFlags(set *pflag.FlagSet) *CLIFlags {
	cli := &CLIFlags{}
	set.BoolVarP(&cli.Help, "help", "h", false, "Show help message")
	set.BoolVarP(&cli.Version, "version", "V", false, "Print the version and exit")
	set.StringVarP(&cli.Model, "model", "m", "", "Model: haiku, sonnet, opus, fable, or 1 to 4")
	set.StringVarP(&cli.Output, "output", "o", "", "Output mode: apply, print or plan")
	set.BoolVar(&cli.Strict, "strict", false, "Fail if the response contains free text or incomplete code blocks")
	set.StringVarP(&cli.Index, "index", "i", "", "Index: none, public or all, or 0 to 2")
	set.BoolVar(&cli.NoProgress, "no-progress", false, "Hide progress indicator")
	set.BoolVar(&cli.Verbose, "verbose", false, "Print all file paths and the prompt sent to the LLM")
	set.BoolVar(&cli.DryRun, "dry-run", false, "Skip the LLM call, only build and check the request")
	set.BoolVarP(&cli.Batch, "batch", "b", false, "Submit the request as a batch job instead of streaming it")
	return cli
}

// Only flags the user actually set become non-nil, so that prompt file defaults can fill the rest.
func (c *CLIFlags) ToFlaglist(set *pflag.FlagSet) (Flaglist, error) {
	flags := Flaglist{
		Strict:     c.Strict,
		NoProgress: c.NoProgress,
		DryRun:     c.DryRun,
		Batch:      c.Batch,
	}

	if set.Changed("model") {
		model, err := ParseModelName(c.Model)
		if err != nil {
			return Flaglist{}, err
		}
		flags.Model = &model
	}
	if set.Changed("output") {
		output, err := ParseOutputName(c.Output)
		if err != nil {
			return Flaglist{}, err
		}
		flags.Output = &output
	}
	if set.Changed("index") {
		index, err := ParseIndexSetting(c.Index)
		if err != nil {
			return Flaglist{}, err
		}
		flags.Index = &index
	}
	return flags, nil
}
