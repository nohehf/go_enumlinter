package main

// Comprehensive test cases for enum linter

// Float-based enum
type Score float64

const (
	ScoreExcellent Score = 4.5
	ScoreGood      Score = 3.5
	ScoreAverage   Score = 2.5
)

// Bool-based enum
type Flag bool

const (
	FlagEnabled  Flag = true
	FlagDisabled Flag = false
)

// Iota-based enum
type Color int

const (
	ColorRed   Color = iota
	ColorGreen Color = iota
	ColorBlue  Color = iota
)

type Wrapper struct {
	Value Score
}

// Valid functions (no diagnostics expected)
func getExcellentScore() Score {
	return ScoreExcellent
}

func getEnabledFlag() Flag {
	return FlagEnabled
}

func getValidColor() Color {
	return ColorRed
}

// Invalid functions that should fail the linter
func getInvalidScore() Score {
	return 1.0 // want "returning literal '1.0' which is not a valid enum value for type Score"
}

func getInvalidFlag() Flag {
	return true // want "returning 'true' which is not a valid enum value for type Flag"
}

func getInvalidColor() Color {
	return 5 // want "returning literal '5' which is not a valid enum value for type Color"
}

// Test nested functions
func outerFunctionInvalid() Status {
	innerFunction := func() Status {
		return "nested_invalid" // want "returning literal '\"nested_invalid\"' which is not a valid enum value for type Status"
	}
	return innerFunction()
}

// Test multiple return statements
func testMultipleReturnsInvalid() Status {
	if true {
		return StatusActive // valid
	}
	return "invalid" // want "returning literal '\"invalid\"' which is not a valid enum value for type Status"
}

// Test invalid enum constants
func getInvalidEnumConstant2() Priority {
	return 42 // want "returning literal '42' which is not a valid enum value for type Priority"
}
