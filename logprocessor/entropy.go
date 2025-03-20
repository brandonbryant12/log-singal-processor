
package logprocessor

import (
	"math"
)

// <ai_context>
// This file contains the EntropyChangeGenerator, which calculates
// the change in entropy for a specified field between before and after states.
// </ai_context>

type EntropyChangeGenerator struct {
	FieldName string
}

func (ecg *EntropyChangeGenerator) GenerateSignal(logData LogData) float64 {
	beforeVal, ok1 := logData.Before[ecg.FieldName].(string)
	afterVal, ok2 := logData.After[ecg.FieldName].(string)
	if ok1 && ok2 {
		beforeEntropy := calculateEntropy(beforeVal)
		afterEntropy := calculateEntropy(afterVal)
		return afterEntropy - beforeEntropy
	}
	return 0.0
}

func calculateEntropy(s string) float64 {
	if s == "" {
		return 0.0
	}
	freq := make(map[rune]float64)
	for _, r := range s {
		freq[r]++
	}
	entropy := 0.0
	length := float64(len(s))
	for _, count := range freq {
		p := count / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}
      