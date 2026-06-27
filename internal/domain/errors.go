package domain

import "errors"

// ErrNotFound is returned by repositories when a requested record does not exist.
// Handlers should map this to HTTP 404; all other errors map to 500.
var ErrNotFound = errors.New("not found")
