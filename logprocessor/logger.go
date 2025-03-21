package logprocessor

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
)

// Initialize a colored structured logger
var logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
	Level:      slog.LevelInfo,
	TimeFormat: "15:04:05",
	NoColor:    false,
}))

// LogAnomalyInput logs the anomaly input in a compact format using slog
func LogAnomalyInput(input AnomalyInput) {
	// Format the basic identifier as table:column:timestamp
	identifier := fmt.Sprintf("%s:%s:%s",
		input.Table,
		input.Column,
		input.Timestamp.Format("2006-01-02T15:04:05Z07:00"))

	// Create vector string representation
	vectorStrs := make([]string, len(input.SignalVector))
	for i, val := range input.SignalVector {
		name := "unknown"
		if i < len(signalGeneratorNames) {
			name = signalGeneratorNames[i]
		}
		vectorStrs[i] = fmt.Sprintf("%s=%.4f", name, val)
	}

	// Format the before and after values
	beforeStr := fmt.Sprintf("%v", input.BeforeValue)
	afterStr := fmt.Sprintf("%v", input.AfterValue)

	// If the values are long strings, trim them
	const maxValueLength = 30
	if len(beforeStr) > maxValueLength {
		beforeStr = beforeStr[:maxValueLength] + "..."
	}
	if len(afterStr) > maxValueLength {
		afterStr = afterStr[:maxValueLength] + "..."
	}

	// Log the operation, values, and vectors
	logger.Info(input.Operation,
		"id", identifier,
		"before", beforeStr,
		"after", afterStr,
		"signals", strings.Join(vectorStrs, ", "))
}
