package errors_test

import (
	"fmt"
	"strings"

	"github.com/muonsoft/errors"
	"github.com/muonsoft/errors/errorstest"
)

const adminUser = 123

type ForbiddenError struct {
	Action string
	UserID int
}

func (err *ForbiddenError) Error() string {
	return "access denied"
}

// Implement errors.LoggableError interface to set fields into structured logger.
func (err *ForbiddenError) LogFields(logger errors.FieldLogger) {
	logger.SetString("action", err.Action)
	logger.SetInt("userID", err.UserID)
}

func DoSomething(userID int) error {
	if userID != adminUser {
		return errors.Wrap(&ForbiddenError{Action: "DoSomething", UserID: userID})
	}

	return nil
}

func ExampleLog_loggableError() {
	err := DoSomething(1)

	// Log error with structured logger.
	logger := errorstest.NewLogger()
	errors.Log(err, logger)
	fmt.Println(`logged message:`, logger.Message)
	fmt.Println(`logged fields:`, logger.Fields)
	fmt.Printf(
		"logged first line of stack trace: %s %s:%d\n",
		logger.StackTrace[0].Name(),
		logger.StackTrace[0].File()[strings.LastIndex(logger.StackTrace[0].File(), "/")+1:],
		logger.StackTrace[0].Line(),
	)

	// Output:
	// logged message: access denied
	// logged fields: map[action:DoSomething userID:1]
	// logged first line of stack trace: github.com/muonsoft/errors_test.DoSomething example_loggable_test.go:30
}
