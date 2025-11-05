package parser

import (
	"path/filepath"
	"strings"
)

type Language string

const (
	LangGo         Language = "go"
	LangJavaScript Language = "javascript"
	LangTypeScript Language = "typescript"
	LangPython     Language = "python"
	LangJava       Language = "java"
	LangRust       Language = "rust"
	LangUnknown    Language = "unknown"
)


var extensionToLanguage = map[string]Language{
	// Go
	".go": LangGo,

	// JavaScript
	".js":  LangJavaScript,
	".jsx": LangJavaScript,
	".mjs": LangJavaScript,
	".cjs": LangJavaScript,

	// TypeScript
	".ts":  LangTypeScript,
	".tsx": LangTypeScript,

	// Python
	".py":   LangPython,
	".pyw":  LangPython,
	".pyi":  LangPython,

	// Java
	".java": LangJava,

	// Rust
	".rs": LangRust,
}

func DetectLanguage(filePath string) Language {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	if lang, exists := extensionToLanguage[ext]; exists {
		return lang
	}

	return LangUnknown
}

func IsSupported(lang Language) bool {
	return lang != LangUnknown
}

// GetParser returns the appropriate parser for a language
func GetParser(lang Language) Parser {
	switch lang {
	case LangGo:
		return NewGoParser()
	case LangJavaScript, LangTypeScript:
		return NewJSParser()
	case LangPython:
		return NewPythonParser()
	default:
		return NewGenericParser() // Fallback: regex-based
	}
}
