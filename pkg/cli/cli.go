package cli

import (
	"fmt"
	"net/http"
	"os"
)

// exit codes
// inspired by sysexits.h
const (
	ExitCodeGeneral        = 1 // Not in sysexits.h, but is standard practice.
	ExitCodeNotFound       = 2
	ExitCodeCancelled      = 3
	ExitCodeForbidden      = 4
	ExitCodeTemporaryError = 5
	ExitCodeUsage          = 64 // EX_USAGE in sysexits.h
)

// Die prints its arguments to stderr, then exits the program with the default
// error code.
func Die(args ...interface{}) {
	DieWithExitCode(ExitCodeGeneral, args...)
}

// DieWithError exits with an error.
// if the error is of the type api.ErrorWithStatusCode, its status will be used,
// otherwise the General exit code will be used
func DieWithError(description string, err error) {
	if httpStatusCodeErr, ok := err.(interface{ HTTPStatusCode() int }); ok {
		switch httpStatusCodeErr.HTTPStatusCode() {
		case http.StatusUnauthorized:
			DieWithExitCode(ExitCodeForbidden, description, err)
			return
		default:
			DieWithExitCode(ExitCodeGeneral, description, err)
		}
	}
	DieWithExitCode(ExitCodeGeneral, description, err)
}

// DieWithExitCode prints its arguments to stderr,
// then exits the program with the given exit code.
func DieWithExitCode(code int, args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(code)
}
