package logprocessor

import (
	"math"
)

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
	freq := make(map[byte]float64)
	bytes := []byte(s)
	for _, b := range bytes {
		freq[b]++
	}
	entropy := 0.0
	length := float64(len(bytes))
	for _, count := range freq {
		p := count / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}
