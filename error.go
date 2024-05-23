package metaser

import (
	"errors"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

type decodeError struct {
	message     string
	fieldErrors field.ErrorList
}

func (de *decodeError) Error() string {
	return de.message
}

// GetErrorList gets field.ErrorList type from underlying error.
func GetErrorList(err error) field.ErrorList {
	de := &decodeError{}
	if errors.As(err, &de) {
		return de.fieldErrors
	}
	return nil
}
