package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sfkleach/re-classify/internal/classifier"
	"github.com/sfkleach/re-classify/internal/config"
)

// Version is set at build time via -ldflags
var Version = "unknown"

func main() {
	// Define command-line flags
	checkOnly := flag.Bool("check", false, "Validate configuration syntax only (don't process input)")
	version := flag.Bool("version", false, "Show version information")

	// Customize usage message
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <config.yaml>\n\n", os.Args[0])
		fmt.Println("re-classify is a token classification tool that uses regex patterns")
		fmt.Println("to classify identifiers and operators in monogram syntax.")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nThe tool reads tokens from stdin (one per line) and outputs")
		fmt.Println("classification results based on the regex patterns in config.yaml")
	}

	// Parse command-line flags
	flag.Parse()

	// Handle version flag
	if *version {
		fmt.Printf("re-classify version %s\n", Version)
		return
	}

	// Check for required config file argument
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Error: exactly one config file must be specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	configFile := args[0]

	// Load configuration
	cfg, err := config.LoadClassifierConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Compile regex patterns
	compiledConfig, err := cfg.CompileRegexes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error compiling regexes: %v\n", err)
		os.Exit(1)
	}

	// If check-only mode, just report success and exit
	if *checkOnly {
		fmt.Println("Configuration syntax is valid")
		return
	}

	// Create classifier engine
	engine := classifier.NewClassifierEngine(compiledConfig)

	// Read tokens from stdin
	scanner := bufio.NewScanner(os.Stdin)
	var tokens []string

	for scanner.Scan() {
		token := strings.TrimSpace(scanner.Text())
		if token != "" {
			tokens = append(tokens, token)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
		os.Exit(1)
	}

	// Build form-start to form-end mappings by analyzing all tokens
	err = engine.BuildFormStartEndMappings(tokens, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building form mappings: %v\n", err)
		os.Exit(1)
	}

	// Process tokens and output classifications
	engine.ProcessTokens(tokens)
}
