package parser

import "io"

// Parser extracts function definitions from source code
type Parser interface {

	// Parse extracts functions from source code
	Parse(reader io.Reader) ([]*Function, error)
}

// Function represents a function/method definition in source code
type Function struct {
	Name      string
	LineStart int
	LineEnd   int
	Type      FunctionType // function, method, closure, etc.
}

type FunctionType string

const (
	TypeFunction FunctionType = "function"
	TypeMethod   FunctionType = "method"
	TypeClosure  FunctionType = "closure"
)

// LEARNING MOMENT: Interval Tree Concept
//
// Problem: Given line number, find containing function in O(log n)
//
// Data Structure: Sorted array of intervals + Binary Search
// Why not a tree? For our use case, sorted array is simpler and fast enough.
//
// Algorithm:
// 1. Sort functions by LineStart
// 2. Binary search to find function where: LineStart <= line <= LineEnd

// FunctionMap provides O(log n) lookup: line number → function
type FunctionMap struct {
	functions []*Function // Sorted by LineStart
}

// NewFunctionMap creates a searchable map of functions
func NewFunctionMap(functions []*Function) *FunctionMap {
	// Sort by line start (ascending)
	// We don't need a sorting algorithm here - Go's sort.Slice does it

	sorted := make([]*Function, len(functions))
	copy(sorted, functions)

	// Bubble sort for educational purposes
	// TODO: replace with sort.Slice in production
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].LineStart > sorted[j].LineStart {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return &FunctionMap{functions: sorted}
}

// FindByLine returns the function containing the given line number
// Time Complexity: O(log n) - Binary search
func (fm *FunctionMap) FindByLine(line int) *Function {
	// Binary search for the function containing this line
	left, right := 0, len(fm.functions)-1

	for left <= right {
		mid := (left + right) / 2
		fn := fm.functions[mid]

		if line < fn.LineStart {

			right = mid - 1
		} else if line > fn.LineEnd {

			left = mid + 1
		} else {

			return fn
		}
	}

	// No function contains this line
	return nil
}

// GetAll returns all functions (for iteration)
func (fm *FunctionMap) GetAll() []*Function {
	return fm.functions
}

// Count returns total number of functions
func (fm *FunctionMap) Count() int {
	return len(fm.functions)
}
