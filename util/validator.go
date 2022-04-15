package util

import (
	"fmt"
)

var Validator = &validator{}

type validator struct {
}

func (*validator) CheckLength(value, name string, min, max int) error {
	length := len(value)
	if length < min || length > max {
		return fmt.Errorf("%s: length should be [%d, %d]", name, min, max)
	}
	return nil
}
