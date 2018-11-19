package rivbin

import (
	"reflect"
	"testing"
)

func TestIsFieldHidden(t *testing.T) {
	var f struct {
		Foo    int
		hidden int
		_      int
		foo    int
		Bar    bool
		Nop    struct {
			a bool
		}
	}
	val := reflect.ValueOf(f)
	expected := []bool{false, true, true, true, false, false}
	for i := 0; i < val.NumField(); i++ {
		if result := isFieldHidden(val, i); expected[i] != result {
			t.Error(i, "unexpected result: isFieldHidden() ==", result)
		}
	}
}
