package commitlog

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
)

// findFile is copied from golang.org/x/tools/cmd/cover/func.go.
// It returns the absolute path of a file given its package relative
// location.
// Ex:commitlog/demo/demo.go -> /Users/bhaines/repo/commitlog/demo/demo.go
func findFile(path string, pack string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	dir, file := filepath.Split(path)
	pkg, err := build.Import(dir, ".", build.FindOnly)
	if err != nil {
		gp := os.Getenv("GOPATH")
		if _, err := os.Stat(filepath.Join(pack, file)); err == nil || !os.IsNotExist(err) {
			return filepath.Join(pack, file), nil
		} else {
			if _, err := os.Stat(filepath.Join(gp, "src", path)); err == nil || !os.IsNotExist(err) {
				return filepath.Join(gp, "src", path), nil
			}
		}
		return "", fmt.Errorf("can't find %q: %v", file, err)
	}
	return filepath.Join(pkg.Dir, file), nil
}

func writeFiles(fileContent map[string][]byte) error {
	for fn, content := range fileContent {
		err := ioutil.WriteFile(fn, content, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

