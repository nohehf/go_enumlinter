package main

// Test cases for enum linter

// String-based enum
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusPending  Status = "pending"
)

// Int-based enum
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
)

// Valid functions that should pass the linter (no diagnostics expected)
func getActiveStatus() Status {
	return StatusActive
}

func getHighPriority() Priority {
	return PriorityHigh
}

// Invalid functions that should fail the linter
func getInvalidStatus() Status {
	return "invalid_status" // want "returning literal '\"invalid_status\"' which is not a valid enum value for type Status"
}

func getInvalidPriority() Priority {
	return 999 // want "returning literal '999' which is not a valid enum value for type Priority"
}

// Test invalid enum constants
func getInvalidEnumConstant() Status {
	return "SomeOtherValue" // want "returning literal '\"SomeOtherValue\"' which is not a valid enum value for type Status"
}

// Test functions that don't return enum types (should be ignored - no diagnostics expected)
func getString() string {
	return "hello"
}

func getInt() int {
	return 42
}

// Test variable declarations
func testVariableDeclarations() {
	// Valid variable declarations (no diagnostics expected)
	var validStatus Status = StatusActive
	var validPriority Priority = PriorityHigh

	// Use the variables to avoid "declared and not used" errors
	_ = validStatus
	_ = validPriority

	// Invalid variable declarations (should fail)
	var invalidStatus Status = "random string" // want "variable 'invalidStatus' assigned literal '\"random string\"' which is not a valid enum value for type Status"
	var invalidPriority Priority = 42          // want "variable 'invalidPriority' assigned literal '42' which is not a valid enum value for type Priority"
	var invalidEnum Status = "SomeOtherValue"  // want "variable 'invalidEnum' assigned literal '\"SomeOtherValue\"' which is not a valid enum value for type Status"

	// Use the invalid variables to avoid "declared and not used" errors
	_ = invalidStatus
	_ = invalidPriority
	_ = invalidEnum
}
