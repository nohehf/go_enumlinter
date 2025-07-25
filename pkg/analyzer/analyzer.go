package analyzer

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "enumlinter",
	Doc:  "Vérifie que seules les valeurs d'énumération sont retournées pour les types d'énumération",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// First pass: collect all type declarations
	typeDecls := make(map[string]*ast.TypeSpec)

	// Map to store enum types and their valid values
	enumTypes := make(map[string]map[string]bool)

	// Map to track which function each return statement belongs to
	returnToFunction := make(map[*ast.ReturnStmt]*ast.FuncDecl)

	inspectTypes := func(node ast.Node) bool {
		genDecl, ok := node.(*ast.GenDecl)

		// Check if the declaration is a `type` to make sure we only inspect user made types
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}

		// Collect all type declarations
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			typeDecls[typeSpec.Name.Name] = typeSpec
		}
		return true
	}

	// Second pass: collect constants and identify enum types
	inspectConsts := func(node ast.Node) bool {

		// Check if the declaration is a `const`
		genDecl, ok := node.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			return true
		}

		// Collect all constants
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// Check if the constant has a type annotation
			typeIdent, ok := valueSpec.Type.(*ast.Ident)
			if !ok {
				continue
			}

			// If this type exists in our type declarations, it might be an enum
			if _, exists := typeDecls[typeIdent.Name]; exists {
				// Initialize the enum values map if it doesn't exist
				if enumTypes[typeIdent.Name] == nil {
					enumTypes[typeIdent.Name] = make(map[string]bool)
				}

				// Add all constants in this spec to the enum values
				for _, name := range valueSpec.Names {
					enumTypes[typeIdent.Name][name.Name] = true
				}
			}
		}
		return true
	}

	// Third pass: collect function declarations and their return statements
	inspectFunctions := func(node ast.Node) bool {
		funcDecl, ok := node.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Find all return statements in this function
		ast.Inspect(funcDecl, func(n ast.Node) bool {
			if returnStmt, ok := n.(*ast.ReturnStmt); ok {
				returnToFunction[returnStmt] = funcDecl
			}
			return true
		})

		return true
	}

	// Fourth pass: check return statements
	inspectReturns := func(node ast.Node) bool {
		returnStmt, ok := node.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		// Get the function that returns the value
		funcDecl := returnToFunction[returnStmt]
		if funcDecl == nil {
			return true
		}

		// Check if the function returns an enum type
		funcType := funcDecl.Type
		if funcType == nil || funcType.Results == nil {
			return true
		}

		hasResults := len(funcType.Results.List) > 0
		if !hasResults {
			return true
		}

		// Get the type of the first result
		typeIdent, ok := funcType.Results.List[0].Type.(*ast.Ident)
		if !ok {
			return true
		}

		// Check if the type is an enum type
		enumValues, exists := enumTypes[typeIdent.Name]
		if !exists || len(enumValues) == 0 {
			return true
		}

		// Check the return value
		for _, result := range returnStmt.Results {
			checkReturnValue(pass, result, enumValues, typeIdent.Name)
		}

		return true
	}

	// Fifth pass: check variable declarations
	inspectVars := func(node ast.Node) bool {
		genDecl, ok := node.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			return true
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// Check if this variable has a type annotation
			if valueSpec.Type == nil {
				continue
			}

			typeIdent, ok := valueSpec.Type.(*ast.Ident)
			if !ok {
				continue
			}

			// Check if the type is an enum type
			enumValues, exists := enumTypes[typeIdent.Name]
			if !exists || len(enumValues) == 0 {
				continue
			}

			// Check each variable in this spec
			for i, name := range valueSpec.Names {
				// Check if there's a corresponding value
				if i < len(valueSpec.Values) {
					checkVariableValue(pass, valueSpec.Values[i], enumValues, typeIdent.Name, name.Name)
				}
			}
		}
		return true
	}

	// Run all inspections
	for _, f := range pass.Files {
		ast.Inspect(f, inspectTypes)
		ast.Inspect(f, inspectConsts)
		ast.Inspect(f, inspectFunctions)
		ast.Inspect(f, inspectReturns)
		ast.Inspect(f, inspectVars)
	}

	return nil, nil
}

func checkReturnValue(pass *analysis.Pass, result ast.Expr, enumValues map[string]bool, enumTypeName string) {
	switch v := result.(type) {
	case *ast.Ident:
		// Check if it's a valid enum constant
		if !enumValues[v.Name] {
			pass.Reportf(result.Pos(), "returning '%s' which is not a valid enum value for type %s", v.Name, enumTypeName)
		}
	case *ast.BasicLit:
		// Any literal (string, int, etc.) is not allowed for enum types
		pass.Reportf(result.Pos(), "returning literal '%s' which is not a valid enum value for type %s", v.Value, enumTypeName)
	}
}

func checkVariableValue(pass *analysis.Pass, value ast.Expr, enumValues map[string]bool, enumTypeName string, varName string) {
	switch v := value.(type) {
	case *ast.Ident:
		// Check if it's a valid enum constant
		if !enumValues[v.Name] {
			pass.Reportf(value.Pos(), "variable '%s' assigned '%s' which is not a valid enum value for type %s", varName, v.Name, enumTypeName)
		}
	case *ast.BasicLit:
		// Any literal (string, int, etc.) is not allowed for enum types
		pass.Reportf(value.Pos(), "variable '%s' assigned literal '%s' which is not a valid enum value for type %s", varName, v.Value, enumTypeName)
	}
}
