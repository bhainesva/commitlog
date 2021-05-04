package commitlog

import (
	"bytes"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/token"
	"strings"
	"testing"
)

func TestRemoveDeadCode(t *testing.T) {
	tests := []struct{
		filename string
		codeVersions []string
	}{
		{
			filename: "filename.go",
			codeVersions: []string{
				`package main

import "fmt"

type Person struct {
	Name     string
	Unused   LayeredUnused
}

type LayeredUnused string

func main() {
	fmt.Println(Person{Name: "Hi"})
}`,
`package main

import "fmt"

type Person struct {
	Name string
}

type LayeredUnused string

func main() {
	fmt.Println(Person{Name: "Hi"})
}`,
`package main

import "fmt"

type Person struct {
	Name string
}

func main() {
	fmt.Println(Person{Name: "Hi"})
}`,
			},
		},
	}

	for _, test := range tests {
		fset := token.NewFileSet()
		d := decorator.NewDecorator(fset)
		dstree, err := d.Parse(test.codeVersions[0])
		if err != nil {
			t.Error("unable to parse initial test code: ", err)
		}
		previousTrees := map[string]*dst.File{test.filename: dstree}

		for i:=0;i<len(test.codeVersions);i++ {
			prunedTrees, changed, err := removeDeadCode(previousTrees, fset, map[string]*decorator.Decorator{test.filename: d})
			previousTrees = prunedTrees
			if err != nil {
				t.Error("error removing dead code", err)
			}
			if i == len(test.codeVersions) - 1 {
				if changed {
					t.Errorf("expected no changes on iteration %d", i+1)
				}
			} else {
				if !changed {
					t.Errorf("expected no changes on iteration %d", i+1)
				}

				prunedTree := prunedTrees[test.filename]
				var out []byte
				buf := bytes.NewBuffer(out)
				r := decorator.NewRestorer()
				r.Fprint(buf, prunedTree)
				if strings.TrimSpace(buf.String()) != test.codeVersions[i+1] {
					t.Errorf("expected pass %d of pruned code to be:\n%s\nbut got:\n%s", i+1, test.codeVersions[i+1], buf.String())
				}
			}
		}
	}
}
