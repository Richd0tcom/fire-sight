package parser

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// LEARNING MOMENT: Python's Indentation-Based Syntax
//
// Unlike C-style languages with braces, Python uses indentation to define scope.
//
// Challenge: Determine where a function ends
// Solution: Track indentation levels
//
// Example:
//   def foo():        # indent = 0, function starts
//       x = 1         # indent = 4, inside function
//       if x > 0:     # indent = 4, still inside
//           y = 2     # indent = 8, inside if
//       return x      # indent = 4, back to function level
//   def bar():        # indent = 0, new function (foo ended)

type PythonParser struct {
	funcPattern *regexp.Regexp
}

func NewPythonParser() *PythonParser {
	// Pattern: def function_name( ... ): or async def function_name( ... ):
	pattern := regexp.MustCompile(`^(\s*)(async\s+)?def\s+(\w+)\s*\(`)
	
	return &PythonParser{funcPattern: pattern}
}

func (pp *PythonParser) Parse(reader io.Reader) ([]*Function, error) {
	scanner := bufio.NewScanner(reader)
	functions := []*Function{}
	
	lineNum := 0
	var currentFunction *Function
	var functionIndent int

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		
		// Skip empty lines and comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Calculate indentation
		indent := pp.getIndentation(line)

		// Check if this line starts a function
		matches := pp.funcPattern.FindStringSubmatch(line)
		if len(matches) >= 4 {
			// Close previous function if exists
			if currentFunction != nil {
				currentFunction.LineEnd = lineNum - 1
				functions = append(functions, currentFunction)
			}

			// Start new function
			fnName := matches[3]
			fnType := TypeFunction
			
			// Check if it's a method (indented = inside class)
			if indent > 0 {
				fnType = TypeMethod
			}

			currentFunction = &Function{
				Name:      fnName,
				LineStart: lineNum,
				Type:      fnType,
			}
			functionIndent = indent
			continue
		}

		// Check if we're inside a function and it ended
		if currentFunction != nil {
			// Function ends when we encounter:
			// 1. A line at same or lower indentation level (except empty/comments)
			// 2. Another function definition
			if indent <= functionIndent && trimmed != "" {
				currentFunction.LineEnd = lineNum - 1
				functions = append(functions, currentFunction)
				currentFunction = nil
			}
		}
	}

	// Close last function if file ended
	if currentFunction != nil {
		currentFunction.LineEnd = lineNum
		functions = append(functions, currentFunction)
	}

	return functions, scanner.Err()
}

// getIndentation counts leading spaces/tabs
func (pp *PythonParser) getIndentation(line string) int {
	indent := 0
	for _, char := range line {
		if char == ' ' {
			indent++
		} else if char == '\t' {
			indent += 4 // Treat tab as 4 spaces
		} else {
			break
		}
	}
	return indent
}