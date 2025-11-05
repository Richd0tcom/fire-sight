package parser

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// GenericParser is a fallback for unsupported languages
// Uses simple heuristics to detect function-like blocks
type GenericParser struct {
	funcPattern *regexp.Regexp
}

func NewGenericParser() *GenericParser {
	// Very generic pattern: word followed by parentheses and brace
	// Catches: func(), method() {, fn: function() {
	pattern := regexp.MustCompile(`(\w+)\s*\([^)]*\)\s*\{`)
	
	return &GenericParser{funcPattern: pattern}
}

func (gp *GenericParser) Parse(reader io.Reader) ([]*Function, error) {
	scanner := bufio.NewScanner(reader)
	functions := []*Function{}
	
	lineNum := 0
	var currentFunction *Function
	braceDepth := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Start function if we detect pattern
		if currentFunction == nil {
			if matches := gp.funcPattern.FindStringSubmatch(line); len(matches) >= 2 {
				currentFunction = &Function{
					Name:      matches[1],
					LineStart: lineNum,
					Type:      TypeFunction,
				}
				braceDepth = 0
			}
		}

		// Track braces to find end
		if currentFunction != nil {
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
			
			if braceDepth <= 0 && strings.Contains(line, "}") {
				currentFunction.LineEnd = lineNum
				functions = append(functions, currentFunction)
				currentFunction = nil
			}
		}
	}

	if currentFunction != nil {
		currentFunction.LineEnd = lineNum
		functions = append(functions, currentFunction)
	}

	return functions, scanner.Err()
}