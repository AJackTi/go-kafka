package command

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidCommand    = errors.New("invalid task command")
	ErrAggregateNotFound = errors.New("task aggregate not found")
	ErrAggregateDeleted  = errors.New("task aggregate is deleted")
	ErrVersionConflict   = errors.New("task aggregate version conflict")
)

// VersionConflictError carries the optimistic-concurrency versions observed
// at the command boundary.
type VersionConflictError struct {
	Expected uint64
	Actual   uint64
}

func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("%v: expected version %d, actual version %d", ErrVersionConflict, e.Expected, e.Actual)
}

func (e *VersionConflictError) Is(target error) bool {
	return target == ErrVersionConflict
}
