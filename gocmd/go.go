// Package gocmd provides a wrapper for interacting with the go cmd line tool
package gocmd

import (
	"bytes"
	"fmt"
	"golang.org/x/tools/cover"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func List() ([]string, error) {
	var stdOut, stdErr bytes.Buffer
	cmd := exec.Command("go", "list", "-find", "all")
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, stdErr.String())
	}
	return strings.Split(stdOut.String(), "\n"), nil
}

func TestCover(pkg, test, coverFilename string) ([]*cover.Profile, error) {
	targetName := pkg
	if strings.HasPrefix(pkg, "/") {
		targetName = "."
	}
	cmd := exec.Command("go", "test", targetName, "-run", "^"+test+"$", "--coverprofile="+coverFilename)
	existingEnv := os.Environ()

	if strings.HasPrefix(pkg, "/") {
		_, modName := modInfo(pkg)
		if modName == "" {
			cmd.Env = append(existingEnv, []string{"GO111MODULE=off"}...)
		}
		cmd.Dir = pkg
	}
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(pkg, "/") {
		coverFilename = filepath.Join(pkg, coverFilename)
	}
	profiles, err := cover.ParseProfiles(coverFilename)
	if err != nil {
		return nil, err
	}
	os.Remove(coverFilename)

	return profiles, nil
}

var (
	slashSlash = []byte("//")
	moduleStr  = []byte("module")
)

// modulePath returns the module path from the gomod file text.
// If it cannot find a module path, it returns an empty string.
// It is tolerant of unrelated problems in the go.mod file.
//
// Copied from cmd/go/internal/modfile.
func modulePath(mod []byte) string {
	for len(mod) > 0 {
		line := mod
		mod = nil
		if i := bytes.IndexByte(line, '\n'); i >= 0 {
			line, mod = line[:i], line[i+1:]
		}
		if i := bytes.Index(line, slashSlash); i >= 0 {
			line = line[:i]
		}
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, moduleStr) {
			continue
		}
		line = line[len(moduleStr):]
		n := len(line)
		line = bytes.TrimSpace(line)
		if len(line) == n || len(line) == 0 {
			continue
		}

		if line[0] == '"' || line[0] == '`' {
			p, err := strconv.Unquote(string(line))
			if err != nil {
				return "" // malformed quoted string or multiline module path
			}
			return p
		}

		return string(line)
	}
	return "" // missing module path
}

func modInfo(dir string) (modDir string, modName string) {
	readModName := func(modFile string) string {
		modBytes, err := ioutil.ReadFile(modFile)
		if err != nil {
			return ""
		}
		return modulePath(modBytes)
	}

	for {
		f := filepath.Join(dir, "go.mod")
		info, err := os.Stat(f)
		if err == nil && !info.IsDir() {
			return dir, readModName(f)
		}

		d := filepath.Dir(dir)
		if len(d) >= len(dir) {
			return "", "" // reached top of file system, no go.mod
		}
		dir = d
	}
}

func TestList(pkg string) ([]string, error) {
	var stdOut, stdErr bytes.Buffer
	targetName := pkg
	if strings.HasPrefix(pkg, "/") {
		targetName = "."
	}
	cmd := exec.Command("go", "test", targetName, "-list", ".*")

	existingEnv := os.Environ()

	if strings.HasPrefix(pkg, "/") {
		_, modName := modInfo(pkg)
		if modName == "" {
			cmd.Env = append(existingEnv, []string{"GO111MODULE=off"}...)
		}
		cmd.Dir = pkg
	}
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, stdErr.String())
	}

	output := strings.Split(stdOut.String(), "\n")
	return filterToTests(output), nil
}

func filterToTests(ss []string) []string {
	var out []string
	for _, s := range ss {
		if s != "" && len(strings.Fields(s)) == 1 && strings.HasPrefix(s, "Test") {
			out = append(out, s)
		}
	}

	return out
}
