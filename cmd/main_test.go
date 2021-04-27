package main

import (
	"fmt"
	"golang.org/x/tools/cover"
	"testing"
)

func TestMergeProfile(t *testing.T) {
	p1 := cover.Profile{
		Blocks: []cover.ProfileBlock{
			cover.ProfileBlock{
				StartLine: 11,
				StartCol:  0,
				EndLine:   13,
				EndCol:    0,
				Count:     0,
			},
			cover.ProfileBlock{
				StartLine: 1,
				StartCol:  0,
				EndLine:   10,
				EndCol:    0,
				Count:     0,
			},
		},
	}

	p2 := cover.Profile{
		Blocks: []cover.ProfileBlock{
			cover.ProfileBlock{
				StartLine: 11,
				StartCol:  0,
				EndLine:   13,
				EndCol:    0,
				Count:     0,
			},
			cover.ProfileBlock{
				StartLine: 1,
				StartCol:  0,
				EndLine:   10,
				EndCol:    0,
				Count:     1,
			},
		},
	}

	//_ := cover.Profile{
	//	Blocks:   []cover.ProfileBlock{
	//		cover.ProfileBlock{
	//			StartLine: 11,
	//			StartCol:  0,
	//			EndLine:   13,
	//			EndCol:    0,
	//			Count:     0,
	//		},
	//		cover.ProfileBlock{
	//			StartLine: 1,
	//			StartCol:  0,
	//			EndLine:   10,
	//			EndCol:    0,
	//			Count:     1,
	//		},
	//	},
	//}

	out := mergeProfiles(&p1, &p2)
	fmt.Println("Length: ", len(out))
	fmt.Printf("%#v", out[0])

	if false {
		t.Errorf("Oops")
	}
}
