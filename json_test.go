package errors_test

import (
	"encoding/json"
	"testing"

	"github.com/muonsoft/errors"
)

func TestStackedError_MarshalJSON(t *testing.T) {
	err := errors.Error("ooh", errors.String("key", "value"))
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
	assertStackRegexp(t, jsonError.StackTrace, JSONStack{
		{
			Function: "github.com/muonsoft/errors_test.TestStackedError_MarshalJSON",
			File:     ".+/errors/json_test.go",
			Line:     11,
		},
	})
	if jsonError.Key != "value" {
		t.Errorf(`expected %#v to have key "key"`, err)
	}
}

func TestWrappedError_MarshalJSON(t *testing.T) {
	err := errors.Wrap(
		errors.Error("ooh", errors.String("deepKey", "deepValue")),
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
	assertStackRegexp(t, jsonError.StackTrace, JSONStack{
		{
			Function: "github.com/muonsoft/errors_test.TestWrappedError_MarshalJSON",
			File:     ".+/errors/json_test.go",
			Line:     39,
		},
	})
	if jsonError.Key != "value" {
		t.Errorf(`expected %#v to have key "key"`, err)
	}
	if jsonError.DeepKey != "deepValue" {
		t.Errorf(`expected %#v to have key "deepKey"`, err)
	}
}

type JSONError struct {
	Error      string    `json:"error"`
	StackTrace JSONStack `json:"stackTrace"`
	Key        string    `json:"key"`
	DeepKey    string    `json:"deepKey"`
}

type JSONStack []JSONFrame

type JSONFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}
