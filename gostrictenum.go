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

		// Extract return types and validate enum return values
		returnTypes := extractReturnTypes(funcType.Results.List)
		validateEnumReturnValues(pass, returnStmt.Results, returnTypes, enumTypes)

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

// extractReturnTypes builds a list of return types from function results,
// handling both named and unnamed return parameters.
func extractReturnTypes(resultsList []*ast.Field) []*ast.Ident {
	var returnTypes []*ast.Ident

	for _, result := range resultsList {
		// Convert the type to an identifier if possible, otherwise nil
		var typeIdent *ast.Ident
		if ident, ok := result.Type.(*ast.Ident); ok {
			typeIdent = ident
		}

		// Handle cases where multiple names share the same type (e.g., func f() (a, b int))
		if len(result.Names) > 0 {
			// Add the type for each named parameter
			for range result.Names {
				returnTypes = append(returnTypes, typeIdent)
			}
		} else {
			// Single unnamed return type
			returnTypes = append(returnTypes, typeIdent)
		}
	}

	return returnTypes
}

// validateEnumReturnValues checks each return value against its corresponding return type
// and reports violations when non-enum values are returned for enum types.
func validateEnumReturnValues(pass *analysis.Pass, returnValues []ast.Expr, returnTypes []*ast.Ident, enumTypes map[string]map[string]bool) {
	for i, returnValue := range returnValues {
		// Ensure we don't go out of bounds (shouldn't happen in valid Go)
		if i >= len(returnTypes) {
			break
		}

		typeIdent := returnTypes[i]
		if typeIdent == nil {
			continue // Not an identifier type (e.g., interface{}, *int, etc.)
		}

		// Check if this return type is an enum type
		enumValues, isEnumType := enumTypes[typeIdent.Name]
		if !isEnumType || len(enumValues) == 0 {
			continue // Not an enum type
		}

		// Validate this return value against the enum type
		checkReturnValue(pass, returnValue, enumValues, typeIdent.Name)
	}
}
