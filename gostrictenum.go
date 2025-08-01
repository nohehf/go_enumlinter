package gostrictenum

import (
	"go/ast"
	"go/token"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

type Settings struct{}

type GoStrictEnumLinter struct {
	settings Settings
}

func init() {
	register.Plugin("gostrictenum", New)
}

// https://github.com/golangci/example-plugin-module-linter/blob/main/example.go
func New(settings any) (register.LinterPlugin, error) {
	// The configuration type will be map[string]any or []interface, it depends on your configuration.
	// You can use https://github.com/go-viper/mapstructure to convert map to struct.

	s, err := register.DecodeSettings[Settings](settings)
	if err != nil {
		return nil, err
	}

	return &GoStrictEnumLinter{settings: s}, nil
}

func (l *GoStrictEnumLinter) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		{
			Name: "gostrictenum",
			Doc:  "Check that only enum values are returned for enum types",
			Run:  l.run,
		},
	}, nil
}

func (l *GoStrictEnumLinter) GetLoadMode() string {
	// NOTE: the mode can be `register.LoadModeSyntax` or `register.LoadModeTypesInfo`.
	// - `register.LoadModeSyntax`: if the linter doesn't use types information.
	// - `register.LoadModeTypesInfo`: if the linter uses types information.

	return register.LoadModeSyntax
}

func (f *GoStrictEnumLinter) run(pass *analysis.Pass) (interface{}, error) {
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

		// Build a list of return types to match with return values by position
		returnTypes := make([]*ast.Ident, 0)
		for _, result := range funcType.Results.List {
			// Handle cases where multiple names share the same type (e.g., func f() (a, b int))
			if len(result.Names) > 0 {
				// Multiple names for this type
				for range result.Names {
					if typeIdent, ok := result.Type.(*ast.Ident); ok {
						returnTypes = append(returnTypes, typeIdent)
					} else {
						returnTypes = append(returnTypes, nil)
					}
				}
			} else {
				// Single unnamed return type
				if typeIdent, ok := result.Type.(*ast.Ident); ok {
					returnTypes = append(returnTypes, typeIdent)
				} else {
					returnTypes = append(returnTypes, nil)
				}
			}
		}

		// Check each return value against its corresponding return type
		for i, returnValue := range returnStmt.Results {
			if i >= len(returnTypes) {
				break // More return values than types (shouldn't happen in valid Go)
			}

			typeIdent := returnTypes[i]
			if typeIdent == nil {
				continue // Not an identifier type
			}

			// Check if the type is an enum type
			enumValues, exists := enumTypes[typeIdent.Name]
			if !exists || len(enumValues) == 0 {
				continue // Not an enum type
			}

			// Check this specific return value against its corresponding enum type
			checkReturnValue(pass, returnValue, enumValues, typeIdent.Name)
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

func checkReturnValue(pass *analysis.Pass, result ast.Expr, enumValues map[string]bool, enumTypeName string) bool {
	switch v := result.(type) {
	case *ast.Ident:
		// Check if it's a valid enum constant
		if !enumValues[v.Name] {
			pass.Reportf(result.Pos(), "returning '%s' which is not a valid enum value for type %s", v.Name, enumTypeName)
			return false
		}
	case *ast.BasicLit:
		// Any literal (string, int, etc.) is not allowed for enum types
		pass.Reportf(result.Pos(), "returning literal '%s' which is not a valid enum value for type %s", v.Value, enumTypeName)
		return false
	}
	return true
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
