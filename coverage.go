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

// mergeProfiles combines a set of new coverage profiles into a set of existing
// coverage profiles. A profile block that is covered in any of the input profiles
// will be covered in the output profiles. The function returns the new profiles, and
// a number representing the net lines covered added by the new profiles.
// This function makes some assumptions about the profiles. First that the codeblocks between
// the different profiles have the same positions in the code. Second that multiple new profiles
// will not be provided for the same file. map[string]*cover.Profile might be a more representative
// type, but less convenient. The Profiles/Blocks in the output are not necessarily ordered.
func mergeProfiles(existingProfiles, newProfiles []*cover.Profile) ([]*cover.Profile, int) {
	type blockPos struct {
		SCol, ECol, SLine, ELine int
	}
	var (
		coverageGain           = 0
		outProfiles            []*cover.Profile
		existingProfilesByFile = map[string][]*cover.Profile{}
		newProfilesByFile      = map[string][]*cover.Profile{}
		filenames              = map[string]struct{}{}
	)

	for _, profile := range existingProfiles {
		existingProfilesByFile[profile.FileName] = append(existingProfilesByFile[profile.FileName], profile)
		filenames[profile.FileName] = struct{}{}
	}

	for _, profile := range newProfiles {
		newProfilesByFile[profile.FileName] = append(newProfilesByFile[profile.FileName], profile)
		filenames[profile.FileName] = struct{}{}
	}

	for file, _ := range filenames {
		blockByPos := map[blockPos]*cover.ProfileBlock{}
		existingProfiles := existingProfilesByFile[file]
		for _, profile := range existingProfiles {
			for _, block := range profile.Blocks {
				block := block
				pos := blockPos{SCol: block.StartCol, ECol: block.EndCol, SLine: block.StartLine, ELine: block.EndLine}
				blockByPos[pos] = &block
			}
		}

		newProfiles := newProfilesByFile[file]
		for _, profile := range newProfiles {
			for _, block := range profile.Blocks {
				block := block
				pos := blockPos{SCol: block.StartCol, ECol: block.EndCol, SLine: block.StartLine, ELine: block.EndLine}
				if existingBlock, ok := blockByPos[pos]; ok {
					if block.Count == 1 {
						if existingBlock.Count == 0 {
							coverageGain += 1 + block.EndLine - block.StartLine
							existingBlock.Count = 1
						}
					}
				} else {
					blockByPos[pos] = &block
					if block.Count == 1 {
						coverageGain += 1 + block.EndLine - block.StartLine
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

// uncoveredCodeDeletingApplication provides the context and helper methods
// needed to use dst.Apply to traverse a DST and remove uncovered code
type uncoveredCodeDeletingApplication struct {
	fset     *token.FileSet
	profile  *cover.Profile
	m        decorator.Map
	toDelete map[dst.Node]struct{}
}

func (u *uncoveredCodeDeletingApplication) pre(cursor *dstutil.Cursor) bool {
	node := cursor.Node()
	astNode := u.m.Ast.Nodes[node]
	// I don't know why the node might be non-nil while the astNode is nil,
	// but it happens
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

		u.toDelete[cursor.Parent()] = struct{}{}
		return false
	}
	return true
}

func (u *uncoveredCodeDeletingApplication) post(cursor *dstutil.Cursor) bool {
	if cursor.Node() == nil {
		return true
	}
	if _, ok := u.toDelete[cursor.Node()]; ok {
		if cursor.Index() >= 0 {
			cursor.Delete()
			return true
		}
	}
	return true
}

// constructCoveredDSTs constructs DSTs from a list of code coverage profiles. It returns the DSTs in a map keyed by the
// absolute filepath of the profiled file. It also returns a map of decorators with the same keys, and the fileset used.
func constructCoveredDSTs(profiles []*cover.Profile) (map[string]*dst.File, *token.FileSet, map[string]*decorator.Decorator, error) {
	var (
		files = map[string]*dst.File{}
		fset = token.NewFileSet()
		decorators = map[string]*decorator.Decorator{}
	)

	for _, profile := range profiles {
		p, err := findFile(profile.FileName)
		if err != nil {
			return nil, nil, nil, err
		}

		d := decorator.NewDecorator(fset)
		dstFile, err := d.ParseFile(p, nil, parser.ParseComments)
		if err != nil {
			return nil, nil, nil, err
		}

		tree, err := constructCoveredDST(fset, profile, dstFile, d)
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

// constructCoveredDST constructs a DST containing the contents of the covered/untracked portions
// of a code coverage profile. It also returns the decorator used, in case the caller needs to
// reference the Dst/Ast maps it contains.
func constructCoveredDST(fset *token.FileSet, profile *cover.Profile, dstFile *dst.File, dec *decorator.Decorator) (*dst.File, error) {
	application := uncoveredCodeDeletingApplication{
		fset:     fset,
		profile:  profile,
		m:        dec.Map,
		toDelete: map[dst.Node]struct{}{},
	}
	newTree := dstutil.Apply(dstFile, application.pre, application.post).(*dst.File)
	return newTree, nil
}
