package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sfkleach/regexptable"
	"gopkg.in/yaml.v3"
)

// Pre-compiled regex for detecting non-zero substitution variables
var nonZeroSubstRegex = regexp.MustCompile(`\$[1-9]`)

// SurroundRegexpConfig represents a start/endings pair with regex substitution
type SurroundRegexpConfig struct {
	Start   string   `yaml:"start"`
	End     string   `yaml:"end"`
	Endings []string `yaml:"endings"`
}

// OperatorConfig represents operator configuration with three precedence values
type OperatorConfig struct {
	Pattern     string   `yaml:"pattern"`
	PrefixPrec  uint16   `yaml:"prefix-prec"`
	InfixPrec   uint16   `yaml:"infix-prec"`
	PostfixPrec uint16   `yaml:"postfix-prec"`
	EndTokens   []string `yaml:"end-tokens,omitempty"` // For form-start tokens
}

// ClassifierConfig represents the configuration structure for the re-classify tool
type ClassifierConfig struct {
	SurroundRegexp      []SurroundRegexpConfig `yaml:"surround-regexp,omitempty"`
	FormPrefixRegexp    []string               `yaml:"form-prefix-regexp,omitempty"`
	SimpleLabelRegexp   []string               `yaml:"simple-label-regexp,omitempty"`
	CompoundLabelRegexp []string               `yaml:"compound-label-regexp,omitempty"`
	VariableRegExp      []string               `yaml:"variable-regexp,omitempty"`

	// Operator configurations with precedence values
	OperatorRegexp []OperatorConfig `yaml:"operator-regexp,omitempty"`
}

// CompiledSurroundRegexp holds a compiled surround regex configuration
type CompiledSurroundRegexp struct {
	StartPattern string   // Original pattern for reference
	EndSubsts    []string // End substitution patterns
}

// StartTokenInfo holds information about a start token including its serial number and endings
type StartTokenInfo struct {
	SerialNumber int             // Serial number for this start/end/endings group
	Endings      map[string]bool // End substitution patterns
}

// CompiledClassifierConfig holds compiled RegexpTable patterns
type CompiledClassifierConfig struct {
	// New efficient start token recognizer - maps start patterns to start token info
	StartTokenTable *regexptable.RegexpTable[*StartTokenInfo] // For quick lookup of serial number and end substitutions
	EndTokenTable   *regexptable.RegexpTable[bool]            // For quick lookup of end tokens mapping to serial numbers

	// All patterns now use RegexpTables for performance
	FormPrefixRegexpTable    *regexptable.RegexpTable[bool]
	SimpleLabelRegexpTable   *regexptable.RegexpTable[bool]
	CompoundLabelRegexpTable *regexptable.RegexpTable[bool]
	VariableRegexpTable      *regexptable.RegexpTable[bool]
	OperatorRegexpTable      *regexptable.RegexpTable[CompiledOperatorConfig]
}

// CompiledOperatorConfig holds a compiled operator configuration
type CompiledOperatorConfig struct {
	PrefixPrec  uint16
	InfixPrec   uint16
	PostfixPrec uint16
	EndTokens   []string
}

// LoadClassifierConfig loads configuration from a YAML file
func LoadClassifierConfig(filename string) (*ClassifierConfig, error) {
	data, err := os.ReadFile(filename) // #nosec G304, this is a CLI application.
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	var config ClassifierConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	return &config, nil
}

// CompileRegexes compiles static regex patterns in the configuration using RegexpTables
// Note: StartTokenTable and EndTokenTable are built dynamically during token analysis
func (cc *ClassifierConfig) CompileRegexes() (*CompiledClassifierConfig, error) {
	// Validate surround-regexp configurations
	for i, surroundConfig := range cc.SurroundRegexp {
		// Ensure that at least one of 'endings' or 'end' is present
		if len(surroundConfig.Endings) == 0 && surroundConfig.End == "" {
			return nil, fmt.Errorf("surround-regexp[%d] must have either 'endings' array or 'end' pattern (or both)", i)
		}

		// Check for invalid backreference usage in endings when end is missing
		if len(surroundConfig.Endings) > 0 && surroundConfig.End == "" {
			for j, ending := range surroundConfig.Endings {
				if nonZeroSubstRegex.MatchString(ending) {
					return nil, fmt.Errorf("surround-regexp[%d].endings[%d] contains backreferences ($1, $2, etc.) but no 'end' pattern is provided for capture groups. Use $0 for the full match or provide an 'end' pattern", i, j)
				}
			}
		}
	}

	compiled := &CompiledClassifierConfig{}
	var err error

	// NOTE: StartTokenTable and EndTokenTable are NOT built here
	// They are built dynamically in BuildFormStartEndMappings based on actual input tokens

	// Build form-prefix-regexp table
	if len(cc.FormPrefixRegexp) > 0 {
		builder := regexptable.NewRegexpTableBuilder[bool]()
		for _, pattern := range cc.FormPrefixRegexp {
			if pattern != "" {
				builder.AddPattern(pattern, true)
			}
		}
		compiled.FormPrefixRegexpTable, err = builder.Build(true, true)
		if err != nil {
			return nil, fmt.Errorf("failed to build form-prefix-regexp table: %w", err)
		}
	}

	// Build simple-label-regexp table
	if len(cc.SimpleLabelRegexp) > 0 {
		builder := regexptable.NewRegexpTableBuilder[bool]()
		for _, pattern := range cc.SimpleLabelRegexp {
			if pattern != "" {
				builder.AddPattern(pattern, true)
			}
		}
		compiled.SimpleLabelRegexpTable, err = builder.Build(true, true)
		if err != nil {
			return nil, fmt.Errorf("failed to build simple-label-regexp table: %w", err)
		}
	}

	// Build compound-label-regexp table
	if len(cc.CompoundLabelRegexp) > 0 {
		builder := regexptable.NewRegexpTableBuilder[bool]()
		for _, pattern := range cc.CompoundLabelRegexp {
			if pattern != "" {
				builder.AddPattern(pattern, true)
			}
		}
		compiled.CompoundLabelRegexpTable, err = builder.Build(true, true)
		if err != nil {
			return nil, fmt.Errorf("failed to build compound-label-regexp table: %w", err)
		}
	}

	// Build variable-regexp table
	if len(cc.VariableRegExp) > 0 {
		builder := regexptable.NewRegexpTableBuilder[bool]()
		for _, pattern := range cc.VariableRegExp {
			if pattern != "" {
				builder.AddPattern(pattern, true)
			}
		}
		compiled.VariableRegexpTable, err = builder.Build(true, true)
		if err != nil {
			return nil, fmt.Errorf("failed to build variable-regexp table: %w", err)
		}
	}

	// Build operator-regexp table
	if len(cc.OperatorRegexp) > 0 {
		builder := regexptable.NewRegexpTableBuilder[CompiledOperatorConfig]()
		for i, opConfig := range cc.OperatorRegexp {
			if opConfig.Pattern != "" {
				compiledOp := CompiledOperatorConfig{
					PrefixPrec:  opConfig.PrefixPrec,
					InfixPrec:   opConfig.InfixPrec,
					PostfixPrec: opConfig.PostfixPrec,
					EndTokens:   opConfig.EndTokens,
				}
				builder.AddPattern(opConfig.Pattern, compiledOp)
			} else {
				return nil, fmt.Errorf("operator-regexp pattern %d is empty", i)
			}
		}
		compiled.OperatorRegexpTable, err = builder.Build(true, true)
		if err != nil {
			return nil, fmt.Errorf("failed to build operator-regexp table: %w", err)
		}
	}

	return compiled, nil
}

// SubstitutePattern performs substitution using capture groups
// groups[0] is the full match ($0), groups[1] is first capture group ($1), etc.
// Also handles $$ as an escape sequence for literal $
func SubstitutePattern(pattern string, groups []string) string {
	if !strings.Contains(pattern, "$") {
		return pattern // Fast path for patterns with no substitutions
	}

	var result strings.Builder
	result.Grow(len(pattern)) // Pre-allocate capacity

	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '$' && i+1 < len(pattern) {
			next := pattern[i+1]
			if next == '$' {
				// Handle $$ -> $
				result.WriteByte('$')
				i++ // Skip the second $
			} else if next >= '0' && next <= '9' {
				// Handle $0, $1, $2, etc.
				groupIndex := int(next - '0')
				if groupIndex < len(groups) {
					result.WriteString(groups[groupIndex])
				} else {
					// Group index out of range, keep original
					result.WriteByte('$')
					result.WriteByte(next)
				}
				i++ // Skip the digit
			} else {
				// Just a $ not followed by digit or $
				result.WriteByte('$')
			}
		} else {
			result.WriteByte(pattern[i])
		}
	}

	return result.String()
}
