package main

import (
	"log"
	"log-signal-processor/dbparsers"
	"log-signal-processor/logprocessor"
	"log-signal-processor/logsimulator"
)

// main is the entry point of the application. It sets up the log parser and signal processor,
// generates mock logs using the default fields, processes them, and prints the anomaly input for each log.
func main() {
	dbType := "oracle"
	targetColumn := "bio" // The column we're analyzing

	// Initialize the appropriate log parser based on the database type
	var parser dbparsers.LogParser
	switch dbType {
	case "oracle":
		parser = &dbparsers.OracleLogParser{}
	case "postgres":
		parser = &dbparsers.PostgresLogParser{}
	default:
		log.Fatal("Unsupported database type")
	}

	// Set up the signal processor with generators for the "bio" field
	processor := logprocessor.SignalProcessor{}
	processor.AddGenerator(&logprocessor.FieldLevenshteinGenerator{FieldName: targetColumn})
	processor.AddGenerator(&logprocessor.EntropyChangeGenerator{FieldName: targetColumn})

	// Generate 5 mock update logs for the "users" table using default fields
	logs := logsimulator.GenerateDefaultLogs(dbType, "UPDATE", "users", 5)

	// Process each log and generate anomaly input
	for _, rawLog := range logs {
		logData, err := parser.ParseLog(rawLog)
		if err != nil {
			log.Printf("Failed to parse log: %v", err)
			continue
		}
		vector := processor.GenerateSignalVector(logData)
		anomalyInput := logprocessor.AnomalyInput{
			Operation:    logData.Operation,
			Table:        logData.Table,
			Column:       targetColumn, // Single column instead of Columns slice
			Timestamp:    logData.Timestamp,
			SignalVector: vector,
		}

		// Use the new slog-based logger instead of pretty printing
		logprocessor.LogAnomalyInput(anomalyInput)
	}
}
