package errors_test

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/muonsoft/errors"
)

func ExampleIs() {
	if _, err := os.Open("non-existing"); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Println("file does not exist")
		} else {
			fmt.Println(err)
		}
	}

	// Output:
	// file does not exist
}

func ExampleAs() {
	if _, err := os.Open("non-existing"); err != nil {
		if pathError, ok := errors.As[*fs.PathError](err); ok {
			fmt.Println("Failed at path:", pathError.Path)
		} else {
			fmt.Println(err)
		}
	}

	// Output:
	// Failed at path: non-existing
}

func ExampleUnwrap() {
	err1 := errors.New("error1")
	err2 := fmt.Errorf("error2: [%w]", err1)
	fmt.Println(err2)
	fmt.Println(errors.Unwrap(err2))

	// Output:
	// error2: [error1]
	// error1
}
