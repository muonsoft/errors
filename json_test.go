package errors_test

import (
	"encoding/json"
	stderrors "errors"
	"testing"

	"github.com/muonsoft/errors"
	"github.com/muonsoft/errors/errorstest"
)

func TestStackedError_MarshalJSON(t *testing.T) {
	err := errors.Errorf("ooh", errors.String("key", "value"))
	jsonData, e := json.Marshal(err)
	if e != nil {
		t.Fatalf("expected %#v to be marshalable into json: %v", err, e)
	}
	var jsonError JSONError
	e = json.Unmarshal(jsonData, &jsonError)
	if e != nil {
		t.Fatalf("failed to unmarshal json: %v", e)
	}

	if jsonError.Error != "ooh" {
		t.Errorf("expected %#v to have error key", err)
	}
	assertStackRegexp(t, jsonError.StackTrace, errorstest.StackTrace{
		{
			Function: "github.com/muonsoft/errors_test.TestStackedError_MarshalJSON",
			File:     ".+/errors/json_test.go",
			Line:     13,
		},
	})
	if jsonError.Key != "value" {
		t.Errorf(`expected %#v to have key "key"`, err)
	}
}

func TestWrappedError_MarshalJSON(t *testing.T) {
	err := errors.Wrap(
		errors.Errorf("ooh", errors.String("deepKey", "deepValue")),
		errors.String("key", "value"),
	)
	jsonData, e := json.Marshal(err)
	if e != nil {
		t.Fatalf("expected %#v to be marshalable into json: %v", err, e)
	}
	var jsonError JSONError
	e = json.Unmarshal(jsonData, &jsonError)
	if e != nil {
		t.Fatalf("failed to unmarshal json: %v", e)
	}

	if jsonError.Error != "ooh" {
		t.Errorf("expected %#v to have error key", err)
	}
	assertStackRegexp(t, jsonError.StackTrace, errorstest.StackTrace{
		{
			Function: "github.com/muonsoft/errors_test.TestWrappedError_MarshalJSON",
			File:     ".+/errors/json_test.go",
			Line:     41,
		},
	})
	if jsonError.Key != "value" {
		t.Errorf(`expected %#v to have key "key"`, err)
	}
	if jsonError.DeepKey != "deepValue" {
		t.Errorf(`expected %#v to have key "deepKey"`, err)
	}
}

func TestJoinedError_MarshalJSON(t *testing.T) {
	err := errors.Join(
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
	jsonData, e := json.Marshal(err)
	if e != nil {
		t.Fatalf("expected %#v to be marshalable into json: %v", err, e)
	}
	var jsonError struct {
		Error      string                `json:"error"`
		StackTrace errorstest.StackTrace `json:"stackTrace"`
		Key1       string                `json:"key1"`
		Key2       string                `json:"key2"`
		Key3       string                `json:"key3"`
		Key4       string                `json:"key4"`
		Key5       string                `json:"key5"`
	}
	e = json.Unmarshal(jsonData, &jsonError)
	if e != nil {
		t.Fatalf("failed to unmarshal json: %v", e)
	}

	if jsonError.Error != "error 1\nerror 2\nerror 3\nerror 4" {
		t.Errorf("expected %#v to have error key", err)
	}
	assertStackRegexp(t, jsonError.StackTrace, errorstest.StackTrace{
		{
			Function: "github.com/muonsoft/errors_test.TestJoinedError_MarshalJSON",
			File:     ".+/errors/json_test.go",
			Line:     74,
		},
	})
	if jsonError.Key1 != "value1" {
		t.Errorf(`expected %#v to have key "key1"`, err)
	}
	if jsonError.Key2 != "value2" {
		t.Errorf(`expected %#v to have key "key2"`, err)
	}
	if jsonError.Key3 != "value3" {
		t.Errorf(`expected %#v to have key "key3"`, err)
	}
	if jsonError.Key4 != "value4" {
		t.Errorf(`expected %#v to have key "key4"`, err)
	}
	if jsonError.Key5 != "value5" {
		t.Errorf(`expected %#v to have key "key5"`, err)
	}
}

type JSONError struct {
	Error      string                `json:"error"`
	StackTrace errorstest.StackTrace `json:"stackTrace"`
	Key        string                `json:"key"`
	DeepKey    string                `json:"deepKey"`
}
