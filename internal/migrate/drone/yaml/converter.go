package yaml

import (
	"github.com/drone/spec/dist/go/convert/drone"
)

func Convert(input []byte) ([]byte, error) {
	b, err := drone.FromBytes(input)
	if err != nil {
		return nil, err
	}
	return b, nil
}
