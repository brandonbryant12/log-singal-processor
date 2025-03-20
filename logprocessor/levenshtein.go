
package logprocessor

// <ai_context>
// This file contains the FieldLevenshteinGenerator, which calculates
// the Levenshtein distance between before and after values of a field.
// </ai_context>

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

// Placeholder for levenshtein distance calculation
func levenshteinDistance(s1, s2 string) int {
	// Implement or import actual function
	return len(s1) + len(s2) // Dummy implementation
}
      