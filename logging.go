package errors

import "errors"

type Logger interface {
	FieldSetter
	Log(message string)
}

func Log(err error, logger Logger) {
	if err == nil {
		return
	}

	for e := err; e != nil; e = errors.Unwrap(e) {
		if s, ok := e.(stackTracer); ok {
			logger.SetStackTrace(s.StackTrace())
		}
		if w, ok := e.(FieldWriter); ok {
			w.WriteFieldsTo(logger)
		}
	}

	logger.Log(err.Error())
}
