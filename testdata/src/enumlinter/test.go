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
