package requester

import (
	"fmt"
	"runtime/debug"
)

type RequesterError struct {
	CallStack   string
	OriginalErr error
}

func (err *RequesterError) Error() string {
	return fmt.Sprintf("%s\n\n%s", err.OriginalErr, err.CallStack)
}

func Error(original error) error {
	return &RequesterError{
		OriginalErr: original,
		CallStack:   fmt.Sprintf("%s", debug.Stack()),
	}
}

func Errorf(format string, args ...interface{}) error {
	return Error(fmt.Errorf(format, args...))
}

func errWrapper(err error) {
	if err != nil {
		// TODO: Send the error by the right channel
	}
}
