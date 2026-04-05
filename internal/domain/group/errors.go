package group

import "errors"

var (
	// ErrGroupNotFound is returned when a group is not found
	ErrGroupNotFound = errors.New("group not found")

	// ErrDuplicateGroupName is returned when a group with the same name already exists for the owner
	ErrDuplicateGroupName = errors.New("group name already exists for this owner")
)
