package errors_test

import (
	stderrors "errors"
	"testing"

	"github.com/muonsoft/errors"
	"github.com/muonsoft/errors/errorstest"
)

func TestLog_noError(t *testing.T) {
	logger := errorstest.NewLogger()

	errors.Log(nil, logger)
}

func TestLog_errorWithoutStack(t *testing.T) {
	logger := errorstest.NewLogger()

	errors.Log(errors.New("ooh"), logger)

	logger.AssertMessage(t, "ooh")
}

func TestLog_errorWithStack(t *testing.T) {
	logger := errorstest.NewLogger()

	err := errors.Wrap(
		errors.Wrap(
			errors.Errorf("ooh", errors.String("deepestKey", "deepestValue")),
			errors.String("deepKey", "deepValue"),
		),
		errors.String("key", "value"),
	)
	errors.Log(err, logger)

	logger.AssertMessage(t, "ooh")
	logger.AssertStackTrace(t, errorstest.StackTrace{
		{
			Function: "github.com/muonsoft/errors_test.TestLog_errorWithStack",
			File:     ".+errors/logging_test.go",
			Line:     30,
		},
	})
	logger.AssertField(t, "key", "value")
	logger.AssertField(t, "deepKey", "deepValue")
	logger.AssertField(t, "deepestKey", "deepestValue")
}

func TestLog_joinedErrors(t *testing.T) {
	logger := errorstest.NewLogger()

	err := errors.Wrap(
		errors.Join(
			errors.Wrap(
				errors.Errorf("error 1", errors.String("key1", "value1")),
				errors.String("key2", "value2"),
			),
			errors.Errorf("error 2", errors.String("key3", "value3")),
			stderrors.Join(
				errors.Errorf("error 3", errors.String("key4", "value4")),
				errors.Errorf("error 4", errors.String("key5", "value5")),
			),
		),
	)
	errors.Log(err, logger)

	logger.AssertMessage(t, "error 1\nerror 2\nerror 3\nerror 4")
	logger.AssertStackTrace(t, errorstest.StackTrace{
		{
			Function: "github.com/muonsoft/errors_test.TestLog_joinedErrors",
			File:     ".+errors/logging_test.go",
			Line:     54,
		},
	})
	logger.AssertField(t, "key1", "value1")
	logger.AssertField(t, "key2", "value2")
	logger.AssertField(t, "key3", "value3")
	logger.AssertField(t, "key4", "value4")
	logger.AssertField(t, "key5", "value5")
}
