package commitlog

import (
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"go/parser"
	"go/token"
	"golang.org/x/tools/cover"
)

type testProfileData map[string][]*cover.Profile

// coveredLines returns a set of line numbers of the
// covered lines from a list of profiles
func coveredLines(pp ...*cover.Profile) map[int]struct{} {
	lines := map[int]struct{}{}

	for _, p := range pp {
		for _, b := range p.Blocks {
			for line := b.StartLine; line <= b.EndLine; line++ {
				lines[line] = struct{}{}
			}
		}
	}

	return lines
}

// numLinesCovered returns the total number of covered lines
// from a set of profiles
func numLinesCovered(pp ...*cover.Profile) int {
	return len(coveredLines(pp...))
}

// inUncoveredBlock returns whether or not the given position
// is in an uncovered segment of the given profile
func inUncoveredBlock(profile *cover.Profile, pos token.Position) bool {
	for _, block := range profile.Blocks {
		if block.Count != 0 {
			continue
		}

		if block.StartLine < pos.Line && pos.Line < block.EndLine {
			return true
		}

		if block.StartLine == pos.Line && block.EndLine == pos.Line {
			if block.StartCol <= pos.Column && pos.Column <= block.EndCol {
				return true
			}
		}

		if block.StartLine == pos.Line {
			if block.StartCol <= pos.Column {
				return true
			}
		}

		if block.EndLine == pos.Line {
			if pos.Column <= block.EndCol {
				return true
			}
		}
	}

	return false
}

func mergeProfiles(profile1, profile2 *cover.Profile) int {
	coverageGain := 0
	type blockPos struct {
		SCol, ECol, SLine, ELine int
	}

	blockByPos := map[blockPos]*cover.ProfileBlock{}
	for _, block := range profile2.Blocks {
		block := block
		pos := blockPos{SCol: block.StartCol, ECol: block.EndCol, SLine: block.StartLine, ELine: block.EndLine}
		if existingBlock, ok := blockByPos[pos]; ok {
			if block.Count == 1 {
				if existingBlock.Count == 1 {
					coverageGain -= block.EndLine - block.StartLine
				} else {
					existingBlock.Count = 1
				}
			}
		} else {
			blockByPos[pos] = &block
			if block.Count == 1 {
				coverageGain += block.EndLine - block.StartLine
			}
		}
	}

	var blocks []cover.ProfileBlock
	for _, block := range blockByPos {
		blocks = append(blocks, *block)
	}

	return coverageGain
}

// TODO: the coverage calculation is suspicious here
// doesn't actually distinguish between existing and new profiles
// makes some assumptions that happen to be true
func mergeAllProfiles(existingProfiles, newProfiles []*cover.Profile) ([]*cover.Profile, int) {
	coverageGain := 0
	type blockPos struct {
		SCol, ECol, SLine, ELine int
	}
	profileByFiles := map[string][]*cover.Profile{}

	for _, profile := range append(existingProfiles, newProfiles...) {
		profileByFiles[profile.FileName] = append(profileByFiles[profile.FileName], profile)
	}

	var outProfiles []*cover.Profile

	for file, fileProfiles := range profileByFiles {
		blockByPos := map[blockPos]*cover.ProfileBlock{}
		for _, profile := range fileProfiles {
			profile := profile
			for _, block := range profile.Blocks {
				block := block
				pos := blockPos{SCol: block.StartCol, ECol: block.EndCol, SLine: block.StartLine, ELine: block.EndLine}
				if existingBlock, ok := blockByPos[pos]; ok {
					if block.Count == 1 {
						if existingBlock.Count == 1 {
							coverageGain -= block.EndLine - block.StartLine
						} else {
							existingBlock.Count = 1
						}
					}
				} else {
					blockByPos[pos] = &block
					if block.Count == 1 {
						coverageGain += block.EndLine - block.StartLine
					}
				}
			}
		}

		var blocks []cover.ProfileBlock
		for _, block := range blockByPos {
			blocks = append(blocks, *block)
		}

		outProfiles = append(outProfiles, &cover.Profile{
			FileName: file,
			Blocks:   blocks,
		})
	}

	return outProfiles, coverageGain
}

type uncoveredCodeDeletingApplication struct {
	fset      *token.FileSet
	profile   *cover.Profile
	m         decorator.Map
	deleteMap map[dst.Node]struct{}
}

func (u *uncoveredCodeDeletingApplication) pre(cursor *dstutil.Cursor) bool {
	node := cursor.Node()
	astNode := u.m.Ast.Nodes[node]
	if node == nil || astNode == nil {
		return false
	}
	pos := astNode.Pos()
	position := u.fset.PositionFor(pos, false)
	if inUncoveredBlock(u.profile, position) {
		if cursor.Index() >= 0 {
			cursor.Delete()
			return false
		}

		u.deleteMap[cursor.Parent()] = struct{}{}
		return true
	}
	return true
}

func (u *uncoveredCodeDeletingApplication) post(cursor *dstutil.Cursor) bool {
	if cursor.Node() == nil {
		return true
	}
	if _, ok := u.deleteMap[cursor.Node()]; ok {
		if cursor.Index() >= 0 {
			cursor.Delete()
			return true
		}
	}
	return true
}

// constructUncoveredDSTs constructs DSTs from a list of code coverage profiles. It returns the DSTs in a map keyed by the
// absolute filepath of the profiled file. It also returns a map of decorators with the same keys, and the fileset used.
func constructUncoveredDSTs(profiles []*cover.Profile) (map[string]*dst.File, *token.FileSet, map[string]*decorator.Decorator, error) {
	files := map[string]*dst.File{}
	decorators := map[string]*decorator.Decorator{}

	fset := token.NewFileSet()
	for _, profile := range profiles {
		tree, d, err := constructUncoveredDST(fset, profile)
		if err != nil {
			return nil, nil, nil, err
		}
		absfp, err := findFile(profile.FileName)
		if err != nil {
			return nil, nil, nil, err
		}
		files[absfp] = tree
		decorators[absfp] = d
	}

	return files, fset, decorators, nil
}

// constructUncoveredDST constructs a DST containing the contents of the covered portion
// of a code coverage profile. It also returns the decorator used, in case the caller needs to
// reference the Dst/Ast maps it contains.
func constructUncoveredDST(fset *token.FileSet, profile *cover.Profile) (*dst.File, *decorator.Decorator, error) {
	p, err := findFile(profile.FileName)
	if err != nil {
		return nil, nil, err
	}

	d := decorator.NewDecorator(fset)
	f, err := d.ParseFile(p, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	application := uncoveredCodeDeletingApplication{
		fset:      fset,
		profile:   profile,
		m:         d.Map,
		deleteMap: map[dst.Node]struct{}{},
	}
	newTree := dstutil.Apply(f, application.pre, application.post).(*dst.File)
	return newTree, d, nil
}
