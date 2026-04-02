package schedule

// Description represents an optional schedule description
type Description struct {
	value string
}

// NewDescription creates a new description
func NewDescription(desc string) Description {
	return Description{value: desc}
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
