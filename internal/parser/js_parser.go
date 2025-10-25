package parser

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// LEARNING MOMENT: Why Regex Instead of AST?
//
// JavaScript/TypeScript parsing is HARD:
// - Multiple syntax styles (ES5, ES6, ESNext)
// - JSX/TSX extensions
// - Complex AST libraries (esprima, babel-parser) are not Go-native
//
// Trade-off: Regex is 90% accurate for function detection, much simpler
//
// Patterns we detect:
// - function foo() {}
// - const foo = () => {}
// - class methods: methodName() {}
// - async function foo() {}

type JSParser struct {
	patterns []*regexp.Regexp
}

func NewJSParser() *JSParser {
	// Compile regex patterns for different function styles
	patterns := []*regexp.Regexp{
		// function declaration: function foo() { or async function foo() {
		regexp.MustCompile(`(?m)^[\s]*(async\s+)?function\s+(\w+)\s*\(`),

		// arrow functions: const foo = () => { or export const foo = async () => {
		regexp.MustCompile(`(?m)^[\s]*(export\s+)?(const|let|var)\s+(\w+)\s*=\s*(async\s*)?\([^)]*\)\s*=>`),

		// class methods: methodName() { or async methodName() {
		regexp.MustCompile(`(?m)^[\s]*(async\s+)?(\w+)\s*\([^)]*\)\s*\{`),

		// object method shorthand: foo() { inside objects
		regexp.MustCompile(`(?m)^\s*(\w+)\s*\([^)]*\)\s*\{`),
	}

	return &JSParser{patterns: patterns}
}

func (jsp *JSParser) Parse(reader io.Reader) ([]*Function, error) {
	scanner := bufio.NewScanner(reader)
	functions := []*Function{}

	lineNum := 0
	var currentFunction *Function
	braceDepth := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		// Check if this line starts a function
		if currentFunction == nil {
			fnName := jsp.detectFunction(line)
			if fnName != "" {
				currentFunction = &Function{
					Name:      fnName,
					LineStart: lineNum,
					Type:      TypeFunction,
				}
				braceDepth = 0
			}
		}

		// Track brace depth to find function end
		if currentFunction != nil {
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")

			if braceDepth <= 0 && strings.Contains(line, "}") {
				// Function ended
				currentFunction.LineEnd = lineNum
				functions = append(functions, currentFunction)
				currentFunction = nil
			}
		}
	}

	// Handle unclosed function (file ended)
	if currentFunction != nil {
		currentFunction.LineEnd = lineNum
		functions = append(functions, currentFunction)
	}

	return functions, scanner.Err()
}

func (jsp *JSParser) detectFunction(line string) string {
	// Try each pattern
	for _, pattern := range jsp.patterns {
		matches := pattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			// Extract function name from capture groups
			// Different patterns have name in different positions
			for i := len(matches) - 1; i >= 2; i-- {
				if matches[i] != "" && !strings.Contains(matches[i], "async") &&
					!strings.Contains(matches[i], "export") &&
					!strings.Contains(matches[i], "const") &&
					!strings.Contains(matches[i], "let") &&
					!strings.Contains(matches[i], "var") {
					return matches[i]
				}
			}
		}
	}
	return ""
}
