package errlist

import "github.com/timmbarton/errors"

var (
	ErrBadRequest = errs.New(errs.ErrCodeBadRequest, 10_0001, "bad request")
)
