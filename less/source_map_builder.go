package less_go

import (
	"fmt"
	"os"
	"strings"
)

type SourceMapBuilder struct {
	options                   SourceMapBuilderOptions
	sourceMap                string
	sourceMapURL             string
	sourceMapInputFilename   string
}

type SourceMapBuilderOptions struct {
	SourceMapFilename              string
	SourceMapURL                   string
	SourceMapOutputFilename        string
	SourceMapInputFilename         string
	SourceMapBasepath              string
	SourceMapRootpath              string
	OutputSourceFiles              bool
	SourceMapGenerator             any
	SourceMapFileInline            bool
	DisableSourcemapAnnotation     bool
}

type SourceMapEnvironment interface {
	EncodeBase64(str string) string
}

type Imports struct {
	Contents              map[string]string
	ContentsIgnoredChars  map[string]int
}

func NewSourceMapBuilder(options SourceMapBuilderOptions) *SourceMapBuilder {
	return &SourceMapBuilder{
		options: options,
	}
}

func (smb *SourceMapBuilder) ToCSS(rootNode SourceMapNode, options map[string]any, imports *Imports, environment SourceMapEnvironment) string {
	sourceMapOutputOptions := SourceMapOutputOptions{
		ContentsIgnoredCharsMap:        imports.ContentsIgnoredChars,
		RootNode:                       rootNode,
		ContentsMap:                    imports.Contents,
		SourceMapFilename:              smb.options.SourceMapFilename,
		SourceMapURL:                   smb.options.SourceMapURL,
		OutputFilename:                 smb.options.SourceMapOutputFilename,
		SourceMapBasepath:              smb.options.SourceMapBasepath,
		SourceMapRootpath:              smb.options.SourceMapRootpath,
		OutputSourceFiles:              smb.options.OutputSourceFiles,
		SourceMapGeneratorConstructor:  func() SourceMapGenerator { 
			// This should be injected from the caller in real usage
			// For now, we'll use a default implementation
			return &defaultSourceMapGenerator{
				mappings: make([]SourceMapMapping, 0),
				sourceContents: make(map[string]string),
			}
		},
	}

	sourceMapOutput := NewSourceMapOutput(sourceMapOutputOptions)
	css := sourceMapOutput.ToCSS(options)
	smb.sourceMap = sourceMapOutput.SourceMap
	smb.sourceMapURL = sourceMapOutput.sourceMapURL
	
	if smb.options.SourceMapInputFilename != "" {
		smb.sourceMapInputFilename = sourceMapOutput.NormalizeFilename(smb.options.SourceMapInputFilename)
	}
	
	if smb.options.SourceMapBasepath != "" && smb.sourceMapURL != "" {
		smb.sourceMapURL = sourceMapOutput.RemoveBasepath(smb.sourceMapURL)
	}
	
	return css + smb.getCSSAppendage(environment)
}

func (smb *SourceMapBuilder) getCSSAppendage(environment SourceMapEnvironment) string {
	sourceMapURL := smb.sourceMapURL
	
	if smb.options.SourceMapFileInline {
		if smb.sourceMap == "" {
			return ""
		}
		sourceMapURL = "data:application/json;base64," + environment.EncodeBase64(smb.sourceMap)
	}

	if smb.options.DisableSourcemapAnnotation {
		return ""
	}

	if sourceMapURL != "" {
		return "/*# sourceMappingURL=" + sourceMapURL + " */"
	}
	return ""
}

func (smb *SourceMapBuilder) GetExternalSourceMap() string {
	return smb.sourceMap
}

func (smb *SourceMapBuilder) SetExternalSourceMap(sourceMap string) {
	smb.sourceMap = sourceMap
}

func (smb *SourceMapBuilder) IsInline() bool {
	return smb.options.SourceMapFileInline
}

func (smb *SourceMapBuilder) GetSourceMapURL() string {
	return smb.sourceMapURL
}

func (smb *SourceMapBuilder) GetOutputFilename() string {
	return smb.options.SourceMapOutputFilename
}

func (smb *SourceMapBuilder) GetInputFilename() string {
	return smb.sourceMapInputFilename
}

// defaultSourceMapGenerator provides a proper implementation for source map generation
type defaultSourceMapGenerator struct {
	mappings       []SourceMapMapping
	sourceContents map[string]string
	sources        []string
	sourceIndexMap map[string]int
}

func (d *defaultSourceMapGenerator) AddMapping(mapping SourceMapMapping) {
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[SourceMapGenerator.AddMapping] gen=%d:%d orig=%d:%d src=%s\n",
			mapping.Generated.Line, mapping.Generated.Column,
			mapping.Original.Line, mapping.Original.Column,
			mapping.Source)
	}
	d.mappings = append(d.mappings, mapping)
	// Track unique sources
	if mapping.Source != "" {
		if d.sourceIndexMap == nil {
			d.sourceIndexMap = make(map[string]int)
		}
		if _, exists := d.sourceIndexMap[mapping.Source]; !exists {
			d.sourceIndexMap[mapping.Source] = len(d.sources)
			d.sources = append(d.sources, mapping.Source)
		}
	}
}

func (d *defaultSourceMapGenerator) SetSourceContent(source, content string) {
	d.sourceContents[source] = content
}

func (d *defaultSourceMapGenerator) ToJSON() map[string]any {
	// Build the mappings string using VLQ encoding
	var mappingsBuilder strings.Builder

	prevGenLine := 0
	prevGenCol := 0
	prevOrigLine := 0
	prevOrigCol := 0
	prevSourceIdx := 0

	// Sort mappings by generated line, then column
	sortedMappings := make([]SourceMapMapping, len(d.mappings))
	copy(sortedMappings, d.mappings)

	// Group mappings by generated line
	lineGroups := make(map[int][]SourceMapMapping)
	maxLine := 0
	for _, m := range sortedMappings {
		line := m.Generated.Line
		if line > maxLine {
			maxLine = line
		}
		lineGroups[line] = append(lineGroups[line], m)
	}

	// Generate mappings for each line
	for line := 1; line <= maxLine; line++ {
		// Add semicolons for empty lines
		if line > prevGenLine+1 {
			for i := prevGenLine + 1; i < line; i++ {
				mappingsBuilder.WriteString(";")
			}
		}

		mappings := lineGroups[line]
		if len(mappings) == 0 {
			if line <= maxLine {
				mappingsBuilder.WriteString(";")
			}
			prevGenLine = line
			prevGenCol = 0
			continue
		}

		// Reset column for new line
		if line > prevGenLine {
			prevGenCol = 0
		}

		for i, m := range mappings {
			if i > 0 {
				mappingsBuilder.WriteString(",")
			}

			// VLQ encode: generated column, source index, original line, original column
			sourceIdx := 0
			if d.sourceIndexMap != nil {
				sourceIdx = d.sourceIndexMap[m.Source]
			}

			// Generated column (relative to previous in same line)
			mappingsBuilder.WriteString(encodeVLQ(m.Generated.Column - prevGenCol))

			// Source index (relative)
			mappingsBuilder.WriteString(encodeVLQ(sourceIdx - prevSourceIdx))

			// Original line (0-indexed, relative)
			mappingsBuilder.WriteString(encodeVLQ((m.Original.Line - 1) - prevOrigLine))

			// Original column (relative)
			mappingsBuilder.WriteString(encodeVLQ(m.Original.Column - prevOrigCol))

			prevGenCol = m.Generated.Column
			prevSourceIdx = sourceIdx
			prevOrigLine = m.Original.Line - 1
			prevOrigCol = m.Original.Column
		}

		mappingsBuilder.WriteString(";")
		prevGenLine = line
	}

	// Build sourcesContent array
	var sourcesContent []any
	for _, src := range d.sources {
		if content, ok := d.sourceContents[src]; ok {
			sourcesContent = append(sourcesContent, content)
		} else {
			sourcesContent = append(sourcesContent, nil)
		}
	}

	result := map[string]any{
		"version":  3,
		"sources":  d.sources,
		"mappings": strings.TrimSuffix(mappingsBuilder.String(), ";"),
	}

	if len(sourcesContent) > 0 {
		result["sourcesContent"] = sourcesContent
	}

	return result
}

// VLQ encoding constants
const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// encodeVLQ encodes a number as a VLQ (Variable Length Quantity)
func encodeVLQ(n int) string {
	var result strings.Builder

	// Handle negative numbers
	var value int
	if n < 0 {
		value = ((-n) << 1) | 1
	} else {
		value = n << 1
	}

	for {
		digit := value & 0x1F // Get the lowest 5 bits
		value >>= 5

		if value > 0 {
			digit |= 0x20 // Set the continuation bit
		}

		result.WriteByte(base64Chars[digit])

		if value == 0 {
			break
		}
	}

	return result.String()
}