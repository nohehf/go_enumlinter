# Enum Linter

A Go static analysis tool that ensures only valid enum values are returned for enum types.

### Examples

#### Valid Enum Usage ✅

```go
type StatusEnum string

const (
    StatusActive   StatusEnum = "active"
    StatusInactive StatusEnum = "inactive"
)

func getStatus() StatusEnum {
    return StatusActive  // ✅ Valid - returns enum constant
}
```

#### Invalid Enum Usage ❌

```go
type StatusEnum string

const (
    StatusActive   StatusEnum = "active"
    StatusInactive StatusEnum = "inactive"
)

func getStatus() StatusEnum {
    return "invalid"  // ❌ Error - returns string literal
}
```

## Features

- **Enum Detection**: Automatically identifies types that have constants defined for them
- **Multi-type Support**: Works with string, int, float, bool, and iota-based enums
- **Return Validation**: Ensures only valid enum constants are returned from functions
- **Comprehensive Testing**: Full test suite using Go's native testing framework

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd go_linter

# Build the linter
go build -o enumlinter cmd/main.go
```

## Usage

### Command Line

```bash
# Analyze a single file
./enumlinter path/to/file.go

# Analyze multiple files
./enumlinter file1.go file2.go

# Analyze a directory
./enumlinter ./path/to/directory
```

## Testing

### Run All Tests

```bash
./run_tests.sh
```

### Run Tests Manually

```bash
# Run analyzer tests
cd pkg/analyzer && go test -v

# Run specific test
go test -v -run TestEnumLinter
```

## Supported Enum Types

- **String-based**: `type Status string`
- **Int-based**: `type Priority int`
- **Float-based**: `type Score float64`
- **Bool-based**: `type Flag bool`
- **Iota-based**: `type Color int` with `iota`

## TODO

Handle a variable being initialized :
```go
var test StatusEnum = "random string" // sadly still works for now ;-;
```

## Development

### Adding New Test Cases

1. Add test files to `testdata/`
2. Use `// want "expected error message"` comments for invalid cases
3. Run tests with `go test -v`

### Extending the Linter

The core logic is in `pkg/analyzer/analyzer.go`. The analyzer:
1. Detects enum types by finding types with constants
2. Validates return statements against enum constants
3. Reports violations with clear error messages

## License

MIT License