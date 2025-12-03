package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	less_go "github.com/toakleaf/less.go/less"
)

const version = "4.2.2-go"

// stringSliceFlag allows for multiple values of a flag (e.g., --include-path multiple times)
type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSliceFlag) Set(value string) error {
	// Support both comma-separated and multiple flag usage
	for _, v := range strings.Split(value, string(os.PathListSeparator)) {
		if v != "" {
			*s = append(*s, v)
		}
	}
	return nil
}

// keyValueFlag for --global-var and --modify-var
type keyValueFlag map[string]string

func (kv *keyValueFlag) String() string {
	pairs := make([]string, 0, len(*kv))
	for k, v := range *kv {
		pairs = append(pairs, k+"="+v)
	}
	return strings.Join(pairs, ", ")
}

func (kv *keyValueFlag) Set(value string) error {
	// Parse "name=value" format
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, expected name=value, got: %s", value)
	}
	(*kv)[parts[0]] = parts[1]
	return nil
}

func main() {
	// Define flags
	var (
		showVersion  bool
		showHelp     bool
		compress     bool
		sourceMap    bool
		sourceMapInline bool
		strictUnits  bool
		jsEnabled    bool
		silent       bool
		mathMode     string
		rewriteUrls  string
		rootpath     string
		urlArgs      string
		includePaths stringSliceFlag
		globalVars   = make(keyValueFlag)
		modifyVars   = make(keyValueFlag)
	)

	// Custom usage message
	flag.Usage = printUsage

	// Boolean flags
	flag.BoolVar(&showVersion, "v", false, "Print version number and exit")
	flag.BoolVar(&showVersion, "version", false, "Print version number and exit")
	flag.BoolVar(&showHelp, "h", false, "Print help and exit")
	flag.BoolVar(&showHelp, "help", false, "Print help and exit")
	flag.BoolVar(&compress, "compress", false, "Compress output CSS")
	flag.BoolVar(&compress, "x", false, "Compress output CSS (shorthand)")
	flag.BoolVar(&sourceMap, "source-map", false, "Generate source map")
	flag.BoolVar(&sourceMapInline, "source-map-inline", false, "Inline source map in CSS output")
	flag.BoolVar(&strictUnits, "strict-units", false, "Enable strict unit checking")
	flag.BoolVar(&jsEnabled, "js", false, "Enable inline JavaScript evaluation")
	flag.BoolVar(&silent, "silent", false, "Suppress output messages")
	flag.BoolVar(&silent, "s", false, "Suppress output messages (shorthand)")

	// String flags
	flag.StringVar(&mathMode, "math", "parens-division", "Math mode: always, parens-division, parens, strict")
	flag.StringVar(&rewriteUrls, "rewrite-urls", "", "URL rewriting: off, local, all")
	flag.StringVar(&rootpath, "rootpath", "", "Set rootpath for URL rewriting")
	flag.StringVar(&urlArgs, "url-args", "", "Query string to append to URLs")

	// Multi-value flags
	flag.Var(&includePaths, "include-path", "Include path for @import (can be specified multiple times, or use OS path separator)")
	flag.Var(&includePaths, "I", "Include path (shorthand)")
	flag.Var(&globalVars, "global-var", "Define a global variable (format: name=value)")
	flag.Var(&modifyVars, "modify-var", "Modify a variable (format: name=value)")

	// Parse flags
	flag.Parse()

	// Handle version and help
	if showVersion {
		fmt.Printf("lessc-go %s (Less Compiler Go Port)\n", version)
		os.Exit(0)
	}

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	// Get positional arguments (input and optional output file)
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: No input file specified")
		printUsage()
		os.Exit(1)
	}

	inputFile := args[0]
	var outputFile string
	if len(args) > 1 {
		outputFile = args[1]
	}

	// Read input (from stdin or file)
	var inputContent []byte
	var absPath string
	var err error

	if inputFile == "-" {
		// Read from stdin
		inputContent, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
		absPath = "stdin"

		// For stdin, use current directory as base path for imports
		cwd, _ := os.Getwd()
		if len(includePaths) == 0 {
			includePaths = append(includePaths, cwd)
		}
	} else {
		// Check if input looks like LESS code (for piped content detection)
		if strings.Contains(inputFile, "{") || strings.Contains(inputFile, "@") {
			// Treat as LESS code directly
			inputContent = []byte(inputFile)
			absPath = "inline"
			cwd, _ := os.Getwd()
			if len(includePaths) == 0 {
				includePaths = append(includePaths, cwd)
			}
		} else {
			// Read from file
			inputContent, err = os.ReadFile(inputFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", inputFile, err)
				os.Exit(1)
			}

			absPath, err = filepath.Abs(inputFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting absolute path: %v\n", err)
				os.Exit(1)
			}

			// Add file's directory to include paths
			fileDir := filepath.Dir(absPath)
			includePaths = append([]string{fileDir}, includePaths...)
		}
	}

	// Build compile options
	options := &less_go.CompileOptions{
		Filename:    absPath,
		Paths:       includePaths,
		Compress:    compress,
		StrictUnits: strictUnits,
	}

	// Enable JavaScript if requested
	if jsEnabled {
		options.EnableJavaScriptPlugins = true
		options.JavascriptEnabled = true
	}

	// Set math mode
	switch strings.ToLower(mathMode) {
	case "always":
		options.Math = less_go.Math.Always
	case "parens-division", "parens_division":
		options.Math = less_go.Math.ParensDivision
	case "parens", "strict":
		options.Math = less_go.Math.Parens
	}

	// Set rewrite URLs mode
	switch strings.ToLower(rewriteUrls) {
	case "all":
		options.RewriteUrls = less_go.RewriteUrls.All
	case "local":
		options.RewriteUrls = less_go.RewriteUrls.Local
	case "off":
		options.RewriteUrls = less_go.RewriteUrls.Off
	}

	if rootpath != "" {
		options.Rootpath = rootpath
	}

	if urlArgs != "" {
		options.UrlArgs = urlArgs
	}

	// Handle global vars
	if len(globalVars) > 0 {
		options.GlobalVars = make(map[string]any)
		for k, v := range globalVars {
			options.GlobalVars[k] = v
		}
	}

	// Handle modify vars
	if len(modifyVars) > 0 {
		options.ModifyVars = make(map[string]any)
		for k, v := range modifyVars {
			options.ModifyVars[k] = v
		}
	}

	// Handle source map options
	if sourceMap || sourceMapInline {
		options.SourceMap = true
		options.SourceMapOptions = &less_go.SourceMapOptions{
			SourceMapFileInline:   sourceMapInline,
			OutputSourceFiles:     true, // Include source content in the source map
		}
		// Set source map filename based on output file
		if outputFile != "" {
			// Use relative path for the sourceMappingURL
			options.SourceMapOptions.SourceMapURL = filepath.Base(outputFile) + ".map"
			options.SourceMapOptions.SourceMapFilename = outputFile + ".map"
			options.SourceMapOptions.SourceMapOutputFilename = outputFile
		}
	}

	// Compile the LESS content
	result, err := less_go.Compile(string(inputContent), options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation error: %v\n", err)
		os.Exit(1)
	}

	css := result.CSS

	// Handle source map output
	var sourceMapContent string
	if sourceMap || sourceMapInline {
		sourceMapContent = result.Map

		if sourceMapInline && sourceMapContent != "" {
			// For inline source maps, the library should have added them
			// If not present, we skip since source map generation needs work
		} else if sourceMap && outputFile != "" && sourceMapContent != "" {
			// Write external source map file
			mapFile := outputFile + ".map"
			if err := os.WriteFile(mapFile, []byte(sourceMapContent), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing source map file %s: %v\n", mapFile, err)
				os.Exit(1)
			}

			// Note: The library already appends the sourceMappingURL comment,
			// so we don't need to add it here

			if !silent {
				fmt.Fprintf(os.Stderr, "Source map written to %s\n", mapFile)
			}
		}
	}

	// Output result
	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(css), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file %s: %v\n", outputFile, err)
			os.Exit(1)
		}
		if !silent {
			fmt.Fprintf(os.Stderr, "Compiled %s -> %s\n", inputFile, outputFile)
		}
	} else {
		// Write to stdout
		writer := bufio.NewWriter(os.Stdout)
		writer.WriteString(css)
		writer.Flush()
	}
}

func printUsage() {
	fmt.Printf(`lessc-go %s (Less Compiler - Go Port)
Usage: lessc-go [options] <input.less|-|"less code"> [output.css]

Input:
  <input.less>       Compile a LESS file
  -                  Read LESS from stdin
  "less code"        Compile LESS code directly (if contains { or @)

Examples:
  lessc-go style.less                      # Output to stdout
  lessc-go style.less style.css            # Output to file
  lessc-go --compress style.less out.css   # Minified output
  cat style.less | lessc-go - out.css      # Read from stdin
  echo "@color: red; .a { color: @color; }" | lessc-go -

Options:
  -h, --help               Print this help message
  -v, --version            Print version number

Compilation:
  -x, --compress           Compress/minify CSS output
  --strict-units           Enable strict unit checking in math operations
  --math=MODE              Math mode: always, parens-division (default), parens
  --js                     Enable inline JavaScript evaluation

Import Paths:
  -I, --include-path=PATH  Add path for @import resolution (repeatable)
                           Can also use OS path separator (: on Unix, ; on Windows)

URL Handling:
  --rewrite-urls=MODE      Rewrite URLs: off, local, all
  --rootpath=PATH          Set root path for URL rewriting
  --url-args=ARGS          Query string to append to all URLs

Variables:
  --global-var=NAME=VALUE  Define a global variable (before parsing)
  --modify-var=NAME=VALUE  Modify a variable (after parsing, overrides)

Source Maps:
  --source-map             Generate external source map (.map file)
  --source-map-inline      Embed source map in CSS output

Output Control:
  -s, --silent             Suppress informational messages

`, version)
}
