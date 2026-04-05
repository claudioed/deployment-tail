package group

import "fmt"

// Description represents an optional group description
type Description struct {
	value string
}

// NewDescription creates a new description with validation
func NewDescription(desc string) (Description, error) {
	if len(desc) > 500 {
		return Description{}, fmt.Errorf("description too long (max 500 characters)")
	}
	return Description{value: desc}, nil
}

// Value returns the underlying string value
func (d Description) Value() string {
	return d.value
}

// String returns the string representation
func (d Description) String() string {
	return d.value
}

// IsEmpty checks if the description is empty
func (d Description) IsEmpty() bool {
	return d.value == ""
}
