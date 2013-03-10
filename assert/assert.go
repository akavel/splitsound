package assert

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func showError(t *testing.T, errmsg ...interface{}) {
	if _, file, line, ok := runtime.Caller(2); ok {
		errmsg = append(errmsg, fmt.Sprintf("(%v:%v)", file, line))
	}
	t.Error(errmsg...)
}

func True(t *testing.T, condition bool, errmsg ...interface{}) {
	if !condition {
		showError(t, errmsg...)
	}
}

func Eq(t *testing.T, expected, got interface{}, errmsg ...interface{}) {
	if !reflect.DeepEqual(expected, got) {
		showError(t, append(errmsg, fmt.Sprintf(`- got: '%v', expected: '%v'`, got, expected))...)
	}
}
