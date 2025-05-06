package balancer

import "errors"

var (
	ErrLimitEnd = errors.New("the request limit has expired")
	ErrDatabase = errors.New("an error occurred while updating the data")
)

