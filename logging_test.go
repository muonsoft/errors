package errors_test

import (
	"testing"

	"github.com/muonsoft/errors"
)

func TestLog_noError(t *testing.T) {
	logger := NewLoggerMock()

	errors.Log(nil, logger)
}

func TestLog_errorWithoutStack(t *testing.T) {
	logger := NewLoggerMock()

	errors.Log(errors.New("ooh"), logger)

	logger.AssertMessage(t, "ooh")
}

func TestLog_errorWithStack(t *testing.T) {
	logger := NewLoggerMock()

	err := errors.Wrap(
		errors.Wrap(
			errors.Errorf("ooh", errors.String("deepestKey", "deepestValue")),
			errors.String("deepKey", "deepValue"),
		),
		errors.String("key", "value"),
	)
	errors.Log(err, logger)

	logger.AssertMessage(t, "ooh")
	logger.AssertStackTrace(t, JSONStack{
		{
			Function: "github.com/muonsoft/errors_test.TestLog_errorWithStack",
			File:     ".+errors/logging_test.go",
			Line:     28,
		},
	})
	logger.AssertField(t, "key", "value")
	logger.AssertField(t, "deepKey", "deepValue")
	logger.AssertField(t, "deepestKey", "deepestValue")
}
