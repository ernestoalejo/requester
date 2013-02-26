package requester

import (
	"fmt"
	"log"
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

func errWrapper(req *Request, err error) {
	if err != nil {
		log.Printf("[%d] An error ocurred !\n", req.Id)
		errLogger.Printf("[%d] %s\n", req.Id, err)
	}
}
