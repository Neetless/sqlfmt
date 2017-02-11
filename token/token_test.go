package token

import (
	"fmt"
	"testing"
)

func TestFileSetPanic(t *testing.T) {
	fs := NewFileSet()

	notProperBase := 0
	assertPanic(t,
		func() { fs.AddFile("test1", notProperBase, 10) },
		fmt.Sprintf("AddFile did not panic by base: %d", notProperBase))
}

func TestFileSetAddFile(t *testing.T) {
	fs := NewFileSet()

	base := 1
	size := 20
	fs.AddFile("test2", base, size)
	expectedBase := 22
	if fs.base != expectedBase {
		t.Errorf("FilsSet fs.base is not proper. Expected: %d, Actual: %d.", expectedBase, fs.base)
	}
}

func assertPanic(t *testing.T, f func(), errState string) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf(errState)
		}
	}()
	f()
}
