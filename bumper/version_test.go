package bumper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRrpmvercmp(t *testing.T) {
	sortedVersionVariants := [][]string{
		{"1.0a"},
		{"1.0b"},
		{"1.0beta"},
		{"1.0p"},
		{"1.0pre"},
		{"1.0rc"},
		{"1.0", "1.00", "01.00"},
		{"1.0.a", "1,0,a", "1;0,a"},
		{"1.0.1", "1,0,1", "1;0,1"},
		{"2", "000002"},
		{"2.0"},
		{"2.1"},
		{"2.1.1"},
		{"2.2"},
		{"3.0"},
		{"3.0.0"},
	}

	// for each list of variants
	for i, currentVariants := range sortedVersionVariants {
		// comparison with lower value variants should yield -1
		for _, lowerVariants := range sortedVersionVariants[:i] {
			assertVersionsComparison(t, lowerVariants, currentVariants, -1)
		}
		// comparison with other variants of the same version should yield 0
		assertVersionsComparison(t, currentVariants, currentVariants, 0)
		// comparison with higher value variants should yield 1
		for _, higherVariants := range sortedVersionVariants[i+1:] {
			assertVersionsComparison(t, higherVariants, currentVariants, 1)
		}
	}
}

// assertVersionsComparison compares each version from aVers with each version from bVers
// and asserts that all those comparisons result in the same expected value.
func assertVersionsComparison(t *testing.T, aVers, bVers []string, expected int) {
	for _, aVer := range aVers {
		for _, bVer := range bVers {
			assert.Equal(t, expected, Rpmvercmp(aVer, bVer), "rpmvercmp(%s, %s) should be %d", aVer, bVer, expected)
		}
	}
}
