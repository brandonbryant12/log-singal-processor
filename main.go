package main

import (
	"fmt"
	"log"
	"log-signal-processor/dbparsers"
	"log-signal-processor/logprocessor"
	"log-signal-processor/logsimulator"
)

// main is the entry point of the application. It sets up the log parser and signal processor,
// generates mock logs using the default fields, processes them, and prints the anomaly input for each log.
func main() {
	dbType := "oracle"

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
	processor.AddGenerator(&logprocessor.FieldLevenshteinGenerator{FieldName: "bio"})
	processor.AddGenerator(&logprocessor.EntropyChangeGenerator{FieldName: "bio"})

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
			Columns:      logData.Columns,
			Timestamp:    logData.Timestamp,
			SignalVector: vector,
		}
		fmt.Printf("AnomalyInput: %+v\n", anomalyInput)
	}
}
