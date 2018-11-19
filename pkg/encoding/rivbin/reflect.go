package rivbin

import (
	"reflect"
	"unicode"
)

func isFieldHidden(val reflect.Value, index int) bool {
	field := val.Type().Field(index)
	if field.Anonymous {
		return true
	}
	for _, r := range field.Name {
		// yes, we do want to return always,
		// we only have a for loop as it seems to be the
		// easiest and most efficient way to do so in Go
		return r == '_' || unicode.IsLower(r)
	}
	return true
}
