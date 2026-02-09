package errors_test

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/muonsoft/errors"
)

const adminUser = 123

type ForbiddenError struct {
	Action string
	UserID int
}

func (err *ForbiddenError) Error() string {
	return "access denied"
}

// Implement errors.LoggableError interface to provide structured attributes for logging.
func (err *ForbiddenError) Attrs() []slog.Attr {
	return []slog.Attr{
		slog.String("action", err.Action),
		slog.Int("userID", err.UserID),
	}
}

func DoSomething(userID int) error {
	if userID != adminUser {
		return errors.Wrap(&ForbiddenError{Action: "DoSomething", UserID: userID})
	}

	return nil
}

func ExampleLog_loggableError() {
	err := DoSomething(1)

	// Get attributes from error
	attrs := errors.Attrs(err)
	
	// Create a map for display
	fields := make(map[string]interface{})
	for _, attr := range attrs {
		fields[attr.Key] = attr.Value.Any()
	}
	
	// Get stack trace
	var stackTrace errors.StackTrace
	for e := err; e != nil; e = errors.Unwrap(e) {
		if s, ok := e.(interface{ StackTrace() errors.StackTrace }); ok {
			stackTrace = s.StackTrace()
			break
		}
	}
	
	fmt.Println(`error message:`, err.Error())
	fmt.Println(`error fields:`, fields)
	fmt.Printf(
		"first line of stack trace: %s %s:%d\n",
		stackTrace[0].Name(),
		stackTrace[0].File()[strings.LastIndex(stackTrace[0].File(), "/")+1:],
		stackTrace[0].Line(),
	)

	// Output:
	// error message: access denied
	// error fields: map[action:DoSomething userID:1]
	// first line of stack trace: github.com/muonsoft/errors_test.DoSomething example_loggable_test.go:32
}
