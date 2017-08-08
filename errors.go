package peg

import (
	"fmt"
)

var (
	errorDismatch            = errorf("the pattern is dismatched")
	errorNotFullMatched      = errorf("the pattern is not full matched")
	errorCornerCase          = errorf("this corner case should never be reached")
	errorCallstackOverflow   = errorf("callstack overflow")
	errorReachedLoopLimit    = errorf("loop limit is reached")
	errorExecuteWhenConsumed = errorf("unable to execute pattern when some text already consumed by caller")
	errorNilConstructor      = errorf("capture constructor is nil")
	errorNilMainPattern      = errorf("the main pattern is nil")

	errorCaseInsensitive = func(name string) error {
		return errorf("case insensitive is not implemented for %q", name)
	}

	errorUndefinedUnicodeRanges = func(name string) error {
		return errorf("unicode class name %q undefined", name)
	}

	errorUndefinedVar = func(name string) error {
		return errorf("variable %q is undefined", name)
	}

	errorInvalidVarName = func(name string) error {
		return errorf("variable name %q is invalid", name)
	}
)

type pegError struct {
	value string
}

func errorf(format string, v ...interface{}) error {
	return &pegError{fmt.Sprintf(format, v...)}
}

func (err *pegError) Error() string {
	return "peg: " + err.value
}
