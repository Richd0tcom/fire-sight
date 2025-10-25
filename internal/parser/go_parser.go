package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
)

type GoParser struct {}


func NewGoParser() *GoParser {
	return &GoParser{}
}

func (gp *GoParser) Parse(reader io.Reader) ([]*Function, error) {

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()

	astFileNode, err:= parser.ParseFile(fset, "", content, parser.ParseComments)

	if err != nil {
		// Syntax error - return empty (don't fail entire analysis)
		return []*Function{}, nil
	}

	functions := []*Function{}

	ast.Inspect(astFileNode, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.FuncDecl:
			fn:= gp.extractFunction(fset, decl)

			functions = append(functions, fn)
		}
		
		return true //continue traversal
	})
	
	return functions, nil
}

func (gp *GoParser) extractFunction(fset *token.FileSet, decl *ast.FuncDecl) *Function {

	// Get line numbers from AST positions
	startPos := fset.Position(decl.Pos())
	endPos := fset.Position(decl.End())

	fnType := TypeFunction
	fnName := decl.Name.Name


	if decl.Recv != nil {
		// Has receiver = method
		fnType = TypeMethod
		// Include receiver in name: "(*User).GetName"
		if len(decl.Recv.List) > 0 {
			receiverType := getTypeName(decl.Recv.List[0].Type)
			fnName = "(" + receiverType + ")." + fnName
		}
	}

	return &Function{
		Name: fnName,
		LineStart: startPos.Line,
		LineEnd: endPos.Line,
		Type: fnType,
	}
}

func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeName(t.X)
	default:
		return "unknown"
	}
}