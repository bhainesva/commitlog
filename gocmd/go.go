package gocmd

import (
	"bytes"
	"fmt"
	"golang.org/x/tools/cover"
	"os"
	"os/exec"
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
	cmd := exec.Command("go", "test", pkg, "-run", "^"+test+"$", "--coverprofile="+coverFilename)
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	profiles, err := cover.ParseProfiles(coverFilename)
	if err != nil {
		return nil, err
	}
	os.Remove(coverFilename)

	return profiles, nil
}

func TestList(pkg string) ([]string, error) {
	var stdOut, stdErr bytes.Buffer
	cmd := exec.Command("go", "test", pkg, "-list", ".*")
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
