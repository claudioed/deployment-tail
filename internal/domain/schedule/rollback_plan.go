package schedule

import "fmt"

// RollbackPlan represents the rollback plan for a deployment schedule
type RollbackPlan struct {
	value string
}

// NewRollbackPlan creates a new RollbackPlan value object with validation
func NewRollbackPlan(plan string) (RollbackPlan, error) {
	if len(plan) > 5000 {
		return RollbackPlan{}, fmt.Errorf("rollback plan too long (max 5000 characters)")
	}

	return RollbackPlan{value: plan}, nil
}

// String returns the string representation of the rollback plan
func (r RollbackPlan) String() string {
	return r.value
}

// IsEmpty checks if the rollback plan is empty
func (r RollbackPlan) IsEmpty() bool {
	return r.value == ""
}

// Equals checks if two rollback plans are equal
func (r RollbackPlan) Equals(other RollbackPlan) bool {
	return r.value == other.value
}
