package commitlog

import (
	"bytes"
	"github.com/dave/dst/decorator"
	"go/token"
	"golang.org/x/tools/cover"
	"reflect"
	"strings"
	"testing"
)

func TestMergeProfiles(t *testing.T) {
	existingProfiles := []*cover.Profile{
		{
			FileName: "File1.go",
			Blocks: []cover.ProfileBlock{
				{StartLine: 0, StartCol: 0, EndLine: 10, EndCol: 10, Count: 1},
				{StartLine: 11, StartCol: 0, EndLine: 22, EndCol: 0, Count: 0},
			},
		},
		{
			FileName: "File2.go",
			Blocks: []cover.ProfileBlock{
				{StartLine: 0, StartCol: 0, EndLine: 10, EndCol: 10, Count: 1},
			},
		},
		{
			FileName: "OnlyInExisting",
			Blocks: []cover.ProfileBlock{
				{StartLine: 0, StartCol: 0, EndLine: 10, EndCol: 10, Count: 1},
			},
		},
	}

	newProfiles := []*cover.Profile{
		{
			FileName: "File1.go",
			Blocks: []cover.ProfileBlock{
				{StartLine: 0, StartCol: 0, EndLine: 10, EndCol: 10, Count: 1},
				{StartLine: 11, StartCol: 0, EndLine: 22, EndCol: 0, Count: 1},
			},
		},
		{
			FileName: "File2.go",
			Blocks: []cover.ProfileBlock{
				{StartLine: 11, StartCol: 0, EndLine: 15, EndCol: 5, Count: 1},
				{StartLine: 16, StartCol: 0, EndLine: 17, EndCol: 0, Count: 1},
			},
		},
		{
			FileName: "OnlyInNew",
			Blocks: []cover.ProfileBlock{
				{StartLine: 0, StartCol: 0, EndLine: 10, EndCol: 10, Count: 1},
			},
		},
	}

	expectedCount := 30
	expectedMergedProfiles := []*cover.Profile{
		{
			FileName: "File1.go",
			Blocks: []cover.ProfileBlock{
				{StartLine: 0, StartCol: 0, EndLine: 10, EndCol: 10, Count: 1},
				{StartLine: 11, StartCol: 0, EndLine: 22, EndCol: 0, Count: 1},
			},
		},
		{
			FileName: "File2.go",
			Blocks: []cover.ProfileBlock{
				{StartLine: 0, StartCol: 0, EndLine: 10, EndCol: 10, Count: 1},
				{StartLine: 11, StartCol: 0, EndLine: 15, EndCol: 5, Count: 1},
				{StartLine: 16, StartCol: 0, EndLine: 17, EndCol: 0, Count: 1},
			},
		},
		{
			FileName: "OnlyInExisting",
			Blocks: []cover.ProfileBlock{
				{StartLine: 0, StartCol: 0, EndLine: 10, EndCol: 10, Count: 1},
			},
		},
		{
			FileName: "OnlyInNew",
			Blocks: []cover.ProfileBlock{
				{StartLine: 0, StartCol: 0, EndLine: 10, EndCol: 10, Count: 1},
			},
		},
	}

	actualMergedProfiles, actualCount := mergeProfiles(existingProfiles, newProfiles)
	actualFiles := map[string]struct{}{}
	expectedFiles := map[string]struct{}{}
	for _, profile := range actualMergedProfiles {
		if _, ok := actualFiles[profile.FileName]; ok {
			t.Errorf("multiple profiles for file: %s, expected on per file", profile.FileName)
		}
		actualFiles[profile.FileName] = struct{}{}
	}

	for _, profile := range expectedMergedProfiles {
		expectedFiles[profile.FileName] = struct{}{}
	}

	if !reflect.DeepEqual(actualFiles, expectedFiles) {
		t.Errorf("coverage profiles did not cover expected Files, found: %v, expected %v", actualFiles, expectedFiles)
	}

	for _, p1 := range actualMergedProfiles {
		for _, p2 := range expectedMergedProfiles {
			if p1.FileName == p2.FileName {
				if !blocksEqual(p1.Blocks, p2.Blocks) {
					t.Errorf("unexpected blocks for file %s, found: %v, expected: %v", p1.FileName, p1.Blocks, p2.Blocks)
				}
			}
		}
	}

	if actualCount != expectedCount {
		t.Errorf("Expected %d new covered lines, got %d", expectedCount, actualCount)
	}
}

func blocksEqual(b1, b2 []cover.ProfileBlock) bool {
	if len(b1) != len(b2) { return false }

	blockSet := map[cover.ProfileBlock]struct{}{}

	for _, b := range b1 {
		blockSet[b] = struct{}{}
	}

	for _, b := range b2 {
		if _, ok := blockSet[b]; !ok {
			return false
		}
	}

	return true
}

func TestConstructCoveredDST(t *testing.T) {
	code := `package main

func uncovered(){
	var a int    // foo
	var b string // bar
}

func partial() string {
	a := "hi"

	if false {
		a += "wow"
	}
	return a
}`
	expectedCode := `package main

func partial() string {
	a := "hi"

	return a
}`

	fset := token.NewFileSet()
	d := decorator.NewDecorator(fset)
	f, err := d.Parse(code)
	if err != nil {
		t.Error("failed to parse sample tree: ", err)
	}

	profile := &cover.Profile{
		FileName: "test/main.go",
		Blocks: []cover.ProfileBlock{
			{ StartLine: 2, StartCol: 0, EndLine: 5, EndCol: 1, Count: 0 },
			{ StartLine: 10, StartCol: 10, EndLine: 12, EndCol: 1, Count: 0 },
		},
	}

	actualDST, err := constructCoveredDST(fset, profile, f, d)
	if err != nil {
		t.Error("err building covered dst: ", err)
	}

	var out []byte
	buf := bytes.NewBuffer(out)
	r := decorator.NewRestorer()
	r.Fprint(buf, actualDST)
	if strings.TrimSpace(buf.String()) != strings.TrimSpace(expectedCode) {
		t.Errorf("Expected file content did not match actual, Expected:\n%s\nActual:\n%s", expectedCode, buf.String())
	}
}