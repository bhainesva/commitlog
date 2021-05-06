package commitlog

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"path/filepath"
)

// findFile is copied from golang.org/x/tools/cmd/cover/func.go.
// It returns the absolute path of a file given its package relative
// location.
// Ex:commitlog/simple/simple.go -> /Users/bhaines/repo/commitlog/simple/simple.go
func findFile(file string) (string, error) {
	dir, file := filepath.Split(file)
	pkg, err := build.Import(dir, ".", build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("can't find %q: %v", file, err)
	}
	return filepath.Join(pkg.Dir, file), nil
}

func WriteFiles(fileContent map[string][]byte) error {
	for fn, content := range fileContent {
		err := ioutil.WriteFile(fn, content, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

