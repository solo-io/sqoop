package storage

import "fmt"

// a special kind of error returned by "create" funcs
// can be used by callers to tell if they can ignore create errors
type alreadyExistsErr struct {
	err error
}

func (err *alreadyExistsErr) Error() string {
	return fmt.Sprintf("already exists: %v", err.err.Error())
}

func NewAlreadyExistsErr(err error) *alreadyExistsErr {
	return &alreadyExistsErr{err: err}
}

func IsAlreadyExists(err error) bool {
	switch err.(type) {
	case *alreadyExistsErr:
		return true
	}
	return false
}
