package schedule

import "errors"

var (
	// ErrScheduleNotFound is returned when a schedule is not found
	ErrScheduleNotFound = errors.New("schedule not found")

	// ErrInvalidScheduleData is returned when schedule data is invalid
	ErrInvalidScheduleData = errors.New("invalid schedule data")

	// ErrScheduleAlreadyExists is returned when trying to create a duplicate schedule
	ErrScheduleAlreadyExists = errors.New("schedule already exists")
)
