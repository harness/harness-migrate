package codeerror

import (
	"fmt"
)

type OpNotSupported struct {
	Name string
}

func (e *OpNotSupported) Error() string {
	return fmt.Sprintf("operation %s not supported", e.Name)
}
