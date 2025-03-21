package main

import (
	"fmt"
	"log"
	"log-signal-processor/cli"
	"log-signal-processor/dbparsers"
	"log-signal-processor/logprocessor"
	"log-signal-processor/logsimulator"
)

// main is the entry point of the application. It sets up the log parser and signal processor,
// generates mock logs using the default fields, processes them, and prints the anomaly input for each log.
func main() {
	// Get configuration from CLI
	config, err := cli.GetConfig()
	if err != nil {
		log.Fatalf("Failed to get configuration: %v", err)
	}

	// Display the selected configuration
	fmt.Printf("Configuration:\n%s\n\n", config)

	// Initialize the appropriate log parser based on the database type
	var parser dbparsers.LogParser
	switch config.DBType {
	case "oracle":
		parser = &dbparsers.OracleLogParser{}
	case "postgres":
		parser = &dbparsers.PostgresLogParser{}
	default:
		log.Fatal("Unsupported database type")
	}

	// Create field configurations from selected fields
	fields := make([]logsimulator.FieldConfig, 0, len(config.SelectedFields))
	for _, fieldName := range config.SelectedFields {
		// Find the matching default field
		for _, defaultField := range logsimulator.GetDefaultFields() {
			if defaultField.Name == fieldName {
				fields = append(fields, defaultField)
				break
			}
		}
	}

	// Generate logs for each selected field
	logs := logsimulator.GenerateLogs(config.DBType, "UPDATE", "users", config.RowCount, fields)

	// Process each log for each selected field
	for _, fieldName := range config.SelectedFields {
		// Create signal processor for this field
		processor := logprocessor.SignalProcessor{}
		processor.AddGenerator(&logprocessor.FieldLevenshteinGenerator{FieldName: fieldName})
		processor.AddGenerator(&logprocessor.EntropyChangeGenerator{FieldName: fieldName})

		fmt.Printf("\n=== Processing field: %s ===\n", fieldName)

		for _, rawLog := range logs {
			logData, err := parser.ParseLog(rawLog)
			if err != nil {
				log.Printf("Failed to parse log: %v", err)
				continue
			}

			vector := processor.GenerateSignalVector(logData)
			beforeValue := logData.Before[fieldName]
			afterValue := logData.After[fieldName]

			anomalyInput := logprocessor.AnomalyInput{
				Operation:    logData.Operation,
				Table:        logData.Table,
				Column:       fieldName,
				Timestamp:    logData.Timestamp,
				BeforeValue:  beforeValue,
				AfterValue:   afterValue,
				SignalVector: vector,
			}

			// Log the anomaly input
			logprocessor.LogAnomalyInput(anomalyInput)
		}
	}
}
