package schedule

import (
	"fmt"
	"strings"
)

// ServiceName represents the name of a service
type ServiceName struct {
	value string
}

// NewServiceName creates a new service name
func NewServiceName(name string) (ServiceName, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ServiceName{}, fmt.Errorf("service name cannot be empty")
	}
	if len(name) > 255 {
		return ServiceName{}, fmt.Errorf("service name cannot exceed 255 characters")
	}
	return ServiceName{value: name}, nil
}

// Value returns the underlying string value
func (s ServiceName) Value() string {
	return s.value
}

// String returns the string representation
func (s ServiceName) String() string {
	return s.value
}
