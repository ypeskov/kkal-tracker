package repositories

import "errors"

var (
	ErrQueryNotFound = errors.New("query not found")
	ErrNotFound      = errors.New("not found")
)
