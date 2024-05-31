package codeerror

import (
	"fmt"
)

type ErrorOpNotSupported struct {
	Name string
}

func (e *ErrorOpNotSupported) Error() string {
	return fmt.Sprintf("operation %s not supported", e.Name)
}
