package logprocessor

type FieldLevenshteinGenerator struct {
	FieldName string
}

func (flg *FieldLevenshteinGenerator) GenerateSignal(logData LogData) float64 {
	beforeVal, ok1 := logData.Before[flg.FieldName].(string)
	afterVal, ok2 := logData.After[flg.FieldName].(string)
	if ok1 && ok2 {
		return float64(levenshteinDistance(beforeVal, afterVal))
	}
	return 0.0
}

// levenshteinDistance calculates the Levenshtein distance between two strings.
// The Levenshtein distance is the minimum number of single-character edits
// (insertions, deletions, or substitutions) required to change one string into the other.
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create the distance matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize the first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
