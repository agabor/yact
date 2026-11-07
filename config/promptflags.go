package config

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/spf13/pflag"
)

// PromptDefaults holds the options a system prompt sets in its flags line. A nil
// pointer means the prompt does not set that option, so a user flag can fill it in.
type PromptDefaults struct {
	Model    *ModelName
	Index    *IndexSetting
	Output   *OutputName
	Strict   bool
	Download bool
}

const promptFlagsPrefix = "flags: "

// Long syntax is recognised by a leading dash, so existing letter-syntax prompt files keep working.
func ParsePromptFlags(name PromptName, content string) (PromptDefaults, error) {
	if !strings.HasPrefix(content, promptFlagsPrefix) {
		return PromptDefaults{}, nil
	}

	firstLine := strings.Split(content, "\n")[0]
	text := strings.TrimSpace(strings.TrimPrefix(firstLine, promptFlagsPrefix))
	if strings.HasPrefix(text, "-") {
		return parseLongFlags(name, text)
	}
	return parseLetterFlags(text), nil
}

func parseLongFlags(name PromptName, text string) (PromptDefaults, error) {
	var modelRaw, indexRaw, outputRaw string
	var defaults PromptDefaults

	set := pflag.NewFlagSet(string(name), pflag.ContinueOnError)
	set.SetOutput(io.Discard)
	set.StringVar(&modelRaw, "model", "", "")
	set.StringVar(&indexRaw, "index", "", "")
	set.StringVar(&outputRaw, "output", "", "")
	set.BoolVar(&defaults.Strict, "strict", false, "")
	set.BoolVar(&defaults.Download, "download", false, "")

	if err := set.Parse(strings.Fields(text)); err != nil {
		return PromptDefaults{}, promptFlagError(name, err)
	}
	if set.NArg() > 0 {
		return PromptDefaults{}, promptFlagError(name, fmt.Errorf("unexpected argument %q", set.Arg(0)))
	}
	if set.Changed("model") {
		model, err := ParseModelName(modelRaw)
		if err != nil {
			return PromptDefaults{}, promptFlagError(name, err)
		}
		defaults.Model = &model
	}
	if set.Changed("index") {
		index, err := ParseIndexSetting(indexRaw)
		if err != nil {
			return PromptDefaults{}, promptFlagError(name, err)
		}
		defaults.Index = &index
	}
	if set.Changed("output") {
		output, err := ParseOutputName(outputRaw)
		if err != nil {
			return PromptDefaults{}, promptFlagError(name, err)
		}
		defaults.Output = &output
	}
	return defaults, nil
}

func promptFlagError(name PromptName, err error) error {
	return fmt.Errorf("prompt '%s' flags: %w", name, err)
}

func parseLetterFlags(text string) PromptDefaults {
	var defaults PromptDefaults
	for _, token := range splitLetterFlags(text) {
		applyLetterFlag(&defaults, token)
	}
	return defaults
}

func applyLetterFlag(defaults *PromptDefaults, token string) {
	if strings.HasPrefix(token, "c") {
		model := modelFromLetterFlag(token)
		defaults.Model = &model
		return
	}

	if strings.HasPrefix(token, "i") {
		if defaults.Index == nil {
			index := indexFromLetterFlag(token)
			defaults.Index = &index
		}
		return
	}

	switch token {
	case "n":
		output := OutputNamePrint
		defaults.Output = &output
	case "p":
		output := OutputNamePlan
		defaults.Output = &output
	case "v":
		defaults.Strict = true
	case "d":
		defaults.Download = true
	}
}

// A digit outside the valid range falls back to sonnet, the default model of the letter syntax.
func modelFromLetterFlag(token string) ModelName {
	if len(token) == 2 {
		if model, err := ParseModelName(token[1:]); err == nil {
			return model
		}
	}
	return ModelSonnet
}

// A digit outside 1 and 2 falls back to public, because index 0 is not a choice a prompt file makes.
func indexFromLetterFlag(token string) IndexSetting {
	if len(token) == 2 {
		if setting, err := ParseIndexSetting(token[1:]); err == nil && setting != IndexNone {
			return setting
		}
	}
	return IndexPublic
}

// A letter followed by a digit is one token, so "c1" is the model 1 and not the letter c.
func splitLetterFlags(text string) []string {
	var tokens []string
	runes := []rune(text)
	for index := 0; index < len(runes); index++ {
		ch := runes[index]
		if (ch == 'c' || ch == 'i') && index+1 < len(runes) && unicode.IsDigit(runes[index+1]) {
			tokens = append(tokens, string(runes[index:index+2]))
			index++
			continue
		}
		tokens = append(tokens, string(ch))
	}
	return tokens
}
