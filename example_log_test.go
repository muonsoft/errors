package errors_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/muonsoft/errors"
)

var (
	// Use errors.New() only for sentinel errors at package level.
	// It would not add a stack trace.
	ErrNotFound = errors.New("not found")
	errSQLError = errors.New("sql error")
)

// To initiate a sentinel error with a stack trace it is recommended to use a
// constructor function and wrap the error with errors.Wrap().
// Use errors.SkipCaller() option to remove constructor function from a stack trace.
func newNotFoundError() error {
	return errors.Wrap(ErrNotFound, errors.SkipCaller())
}

type Product struct {
	ID   int
	Name string
}

type SQLDriver struct {
	err error
}

type Row struct {
	err error
}

func (row *Row) Scan(dest ...interface{}) error {
	return row.err
}

func (driver *SQLDriver) QueryRow(ctx context.Context, sql string, arguments ...interface{}) *Row {
	return &Row{err: driver.err}
}

type ProductRepository struct {
	db *SQLDriver
}

func (repository *ProductRepository) FindByID(ctx context.Context, id int) (*Product, error) {
	const findSQL = `SELECT id, name FROM product WHERE id = ?`

	row := repository.db.QueryRow(ctx, findSQL, id)
	var product Product
	err := row.Scan(&product.ID, &product.Name)
	if errors.Is(err, sql.ErrNoRows) {
		// Error from newNotFoundError will have a stack trace pointing to the line in which
		// it was called.
		return nil, newNotFoundError()
	}
	if err != nil {
		// Use errors.Errorf to wrap the library error with the message context and
		// error fields to be used for structured logging.
		return nil, errors.Errorf(
			"%w: %v", errSQLError, err.Error(),
			errors.String("sql", findSQL),
			errors.Int("productID", id),
		)
	}

	return &product, nil
}

type ErrorJSON struct {
	Error      string                `json:"error"`
	StackTrace []StackTraceFrameJSON `json:"stackTrace"`
	ProductID  int                   `json:"productID"`
	SQL        string                `json:"sql"`
}

type StackTraceFrameJSON struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

func errToJSON(sourceError error) ErrorJSON {
	jsonData, err := json.Marshal(sourceError)
	if err != nil {
		log.Fatal(err)
	}
	var jsonError ErrorJSON
	err = json.Unmarshal(jsonData, &jsonError)
	if err != nil {
		log.Fatal(err)
	}

	return jsonError
}

func ExampleLog_typicalErrorHandling() {
	repository := ProductRepository{db: &SQLDriver{}}

	// Imitating no rows error.
	repository.db.err = sql.ErrNoRows

	_, notFoundError := repository.FindByID(context.Background(), 123)
	if notFoundError != nil {
		// Print error as a text.
		fmt.Println("repository error:", notFoundError)
		fmt.Println("repository error is ErrNotFound:", errors.Is(notFoundError, ErrNotFound))

		// Marshal error into structured JSON.
		jsonError := errToJSON(notFoundError)
		fmt.Println(`repository error as JSON, field "error":`, jsonError.Error)
		fmt.Println(`repository error as JSON, field "stackTrace[0].function":`, jsonError.StackTrace[0].Function)
		fmt.Println(
			`repository error as JSON, field "stackTrace[0].file":`,
			jsonError.StackTrace[0].File[strings.LastIndex(jsonError.StackTrace[0].File, "/")+1:],
		)
		fmt.Println(`repository error as JSON, field "stackTrace[0].line":`, jsonError.StackTrace[0].Line)

		// Get attributes from error
		attrs := errors.Attrs(notFoundError)

		// Get stack trace
		var stackTrace errors.StackTrace
		for e := notFoundError; e != nil; e = errors.Unwrap(e) {
			if s, ok := e.(interface{ StackTrace() errors.StackTrace }); ok {
				stackTrace = s.StackTrace()
				break
			}
		}

		fmt.Println(`log repository error, attrs count:`, len(attrs))
		fmt.Printf(
			"log repository error, first line of stack trace: %s %s:%d\n",
			stackTrace[0].Name(),
			stackTrace[0].File()[strings.LastIndex(stackTrace[0].File(), "/")+1:],
			stackTrace[0].Line(),
		)
	}

	// Imitating driver error.
	repository.db.err = sql.ErrConnDone

	_, sqlError := repository.FindByID(context.Background(), 123)
	if sqlError != nil {
		// Print error as a text.
		fmt.Println("repository error:", sqlError)
		fmt.Println("repository error is errSQLError:", errors.Is(sqlError, errSQLError))

		// Marshal error into structured JSON.
		jsonError := errToJSON(sqlError)
		fmt.Println(`repository error as JSON, field "error":`, jsonError.Error)
		fmt.Println(`repository error as JSON, field "stackTrace[0].function":`, jsonError.StackTrace[0].Function)
		fmt.Println(
			`repository error as JSON, field "stackTrace[0].file":`,
			jsonError.StackTrace[0].File[strings.LastIndex(jsonError.StackTrace[0].File, "/")+1:],
		)
		fmt.Println(`repository error as JSON, field "stackTrace[0].line":`, jsonError.StackTrace[0].Line)

		// Get attributes from error
		attrs := errors.Attrs(sqlError)

		// Create a map for display
		fields := make(map[string]interface{})
		for _, attr := range attrs {
			fields[attr.Key] = attr.Value.Any()
		}

		// Get stack trace
		var stackTrace errors.StackTrace
		for e := sqlError; e != nil; e = errors.Unwrap(e) {
			if s, ok := e.(interface{ StackTrace() errors.StackTrace }); ok {
				stackTrace = s.StackTrace()
				break
			}
		}

		fmt.Println(`log repository error, fields:`, fields)
		fmt.Printf(
			"log repository error, first line of stack trace: %s %s:%d\n",
			stackTrace[0].Name(),
			stackTrace[0].File()[strings.LastIndex(stackTrace[0].File(), "/")+1:],
			stackTrace[0].Line(),
		)
	}

	// Output:
	// repository error: not found
	// repository error is ErrNotFound: true
	// repository error as JSON, field "error": not found
	// repository error as JSON, field "stackTrace[0].function": github.com/muonsoft/errors_test.(*ProductRepository).FindByID
	// repository error as JSON, field "stackTrace[0].file": example_log_test.go
	// repository error as JSON, field "stackTrace[0].line": 62
	// log repository error, attrs count: 0
	// log repository error, first line of stack trace: github.com/muonsoft/errors_test.(*ProductRepository).FindByID example_log_test.go:62
	// repository error: sql error: sql: connection is already closed
	// repository error is errSQLError: true
	// repository error as JSON, field "error": sql error: sql: connection is already closed
	// repository error as JSON, field "stackTrace[0].function": github.com/muonsoft/errors_test.(*ProductRepository).FindByID
	// repository error as JSON, field "stackTrace[0].file": example_log_test.go
	// repository error as JSON, field "stackTrace[0].line": 67
	// log repository error, fields: map[productID:123 sql:SELECT id, name FROM product WHERE id = ?]
	// log repository error, first line of stack trace: github.com/muonsoft/errors_test.(*ProductRepository).FindByID example_log_test.go:67
}
