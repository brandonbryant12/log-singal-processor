package logprocessor

import (
	"time"
)

type LogData struct {
	Operation     string
	Table         string
	RowIdentifier string
	Columns       []string
	Timestamp     time.Time
	Before        map[string]interface{}
	After         map[string]interface{}
}

type SignalGenerator interface {
	GenerateSignal(logData LogData) float64
}

type SignalProcessor struct {
	generators []SignalGenerator
}

func (sp *SignalProcessor) AddGenerator(gen SignalGenerator) {
	sp.generators = append(sp.generators, gen)

	// Register generator name for pretty printing
	var name string
	switch g := gen.(type) {
	case *FieldLevenshteinGenerator:
		name = "Levenshtein(" + g.FieldName + ")"
	case *EntropyChangeGenerator:
		name = "Entropy(" + g.FieldName + ")"
	default:
		name = "Unknown Generator"
	}

	RegisterSignalGenerator(name)
}

func (sp *SignalProcessor) GenerateSignalVector(logData LogData) []float64 {
	vector := make([]float64, len(sp.generators))
	for i, gen := range sp.generators {
		vector[i] = gen.GenerateSignal(logData)
	}
	return vector
}

// GetGenerators returns the list of signal generators
func (sp *SignalProcessor) GetGenerators() []SignalGenerator {
	return sp.generators
}

type AnomalyInput struct {
	Operation    string
	Table        string
	Column       string // Changed from Columns []string to a single Column
	Timestamp    time.Time
	BeforeValue  interface{} // Value of the column before change
	AfterValue   interface{} // Value of the column after change
	SignalVector []float64
}
