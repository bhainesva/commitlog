package commitlog

import (
	"sort"

	"golang.org/x/tools/cover"
)

type testSortingFunction func(testProfileData) []string

// sortHardcodedOrder returns a sorting function that always produces
// the specified ordering
func sortHardcodedOrder(order []string) testSortingFunction {
	return func(testProfileData) []string {
		return order
	}
}

// sortTestsByRawLinesCovered sorts tests by the number of lines they cover
func sortTestsByRawLinesCovered(testProfiles testProfileData) []string {
	var tests []string
	coverageByTest := map[string]int{}

	for test, profiles := range testProfiles {
		tests = append(tests, test)
		coverageByTest[test] = numLinesCovered(profiles...)
	}

	sort.Slice(tests, func(i, j int) bool {
		iCount := coverageByTest[tests[i]]
		jCount := coverageByTest[tests[j]]
		return iCount < jCount
	})
	return tests
}

// sortTestsByNewLinesCovered sorts tests by calculating the number of lines of
// coverage each test would add to the coverage provided by the already sorted
// tests, and selecting the test which provides the smallest number of new lines
func sortTestsByNewLinesCovered(testProfiles testProfileData) []string {
	var (
		sortedTests      []string
		tests            []string
		existingCoverage []*cover.Profile
	)

	for test, _ := range testProfiles {
		tests = append(tests, test)
	}

	for len(tests) > 0 {
		minCoverageGain := -1
		minTestIdx := 0
		var minCoverage []*cover.Profile
		for i, test := range tests {
			profiles := testProfiles[test]
			newCoverage, coverageGain := mergeProfiles(existingCoverage, profiles)
			if minCoverageGain == 0 || coverageGain < minCoverageGain {
				minTestIdx = i
				minCoverageGain = coverageGain
				minCoverage = newCoverage
			}
		}
		sortedTests = append(sortedTests, tests[minTestIdx])
		existingCoverage = minCoverage
		tests = append(tests[:minTestIdx], tests[minTestIdx+1:]...)
	}
	return sortedTests
}

// sortTestsByImportance sorts tests using an 'importance' heuristic
// each line in a file is given a point for every test that covers it
// then tests are ranked by the average value of the lines they cover
func sortTestsByImportance(testProfiles testProfileData) []string {
	var (
		allProfiles []*cover.Profile
		tests       []string
	)

	for test, profiles := range testProfiles {
		tests = append(tests, test)
		allProfiles = append(allProfiles, profiles...)
	}

	lineWeights := scoreLines(allProfiles)

	avgScoreByTest := map[string]float64{}
	for test, profiles := range testProfiles {
		avgScoreByTest[test] = scoreProfiles(profiles, lineWeights)
	}

	sort.Slice(tests, func(i, j int) bool {
		iCount := avgScoreByTest[tests[i]]
		jCount := avgScoreByTest[tests[j]]
		return iCount < jCount
	})

	return tests
}

// scoreLines takes a list of profiles and computes a score for
// each line, which is returned in a map[filename]map[line]score
// the score for a particular line is the number of different profiles
// that cover it
func scoreLines(profiles []*cover.Profile) map[string]map[int]int {
	scores := map[string]map[int]int{}

	for _, profile := range profiles {
		if _, ok := scores[profile.FileName]; !ok {
			scores[profile.FileName] = map[int]int{}
		}

		for _, b := range profile.Blocks {
			if b.Count != 1 {
				continue
			}

			for line := b.StartLine; line <= b.EndLine; line++ {
				scores[profile.FileName][line]++
			}
		}
	}

	return scores
}

// scoreProfiles computes the score for a set of profiles, given a map
// map[filename]map[line]score to pull line scores from
func scoreProfiles(profiles []*cover.Profile, lineWeights map[string]map[int]int) float64 {
	totalScore := 0.0
	totalLines := 0.0

	for _, p := range profiles {
		for _, b := range p.Blocks {
			for line := b.StartLine; line < b.EndLine; line++ {
				totalLines += 1
				totalScore += float64(lineWeights[p.FileName][line])
			}
		}
	}

	if totalLines == 0 {
		return 0
	}

	return totalScore / totalLines
}
