package commitlog

import (
	"golang.org/x/tools/cover"
	"reflect"
	"testing"
)

func TestTestSorts(t *testing.T) {
	profiles := testProfileData{
		"TestOne": []*cover.Profile{
			{
				FileName: "File1.go",
				Blocks: []cover.ProfileBlock{
					{StartLine: 0, StartCol: 0, EndLine: 10, EndCol: 10, Count: 1},
				},
			},
		},
		"TestTwo": []*cover.Profile{
			{
				FileName: "File1.go",
				Blocks: []cover.ProfileBlock{
					{StartLine: 0, StartCol: 0, EndLine: 10, EndCol: 10, Count: 1},
					{StartLine: 11, StartCol: 0, EndLine: 15, EndCol: 5, Count: 1},
				},
			},
		},
		"TestThree": []*cover.Profile{
			{
				FileName: "File1.go",
				Blocks: []cover.ProfileBlock{
					{StartLine: 20, StartCol: 0, EndLine: 26, EndCol: 0, Count: 1},
				},
			},
		},
		"TestFour": []*cover.Profile{
			{
				FileName: "File2.go",
				Blocks: []cover.ProfileBlock{
					{StartLine: 0, StartCol: 0, EndLine: 3, EndCol: 4, Count: 1},
				},
			},
		},
	}

	tests := []struct{
		name string
		sortingFunc testSortingFunction
		expectedOrder []string
	}{
		{
			name: "Harrdcoded sort",
			sortingFunc: sortHardcodedOrder([]string{"TestOne", "TestThree", "TestFour", "TestTwo"}),
			expectedOrder: []string{"TestOne", "TestThree", "TestFour", "TestTwo"},
		},
		{
			name: "Raw lines sort",
			sortingFunc: sortTestsByRawLinesCovered,
			expectedOrder: []string{"TestFour", "TestThree", "TestOne", "TestTwo"},
		},
		{
			name: "New lines sort",
			sortingFunc: sortTestsByNewLinesCovered,
			expectedOrder: []string{"TestFour", "TestThree", "TestOne", "TestTwo"},
		},
		{
			name: "Importance sort",
			sortingFunc: sortTestsByImportance,
			expectedOrder: []string{"TestThree", "TestFour", "TestTwo", "TestOne"},
		},
	}

	for _, test := range tests {
		actualOrder := test.sortingFunc(profiles)
		if !reflect.DeepEqual(actualOrder, test.expectedOrder) {
			t.Errorf("%s: Expected %#v, got %#v", test.name, test.expectedOrder, actualOrder)
		}
	}
}