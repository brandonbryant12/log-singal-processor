package main

import (
	"fmt"
	"log"
	"log-signal-processor/dbparsers"
	"log-signal-processor/logprocessor"
	"log-signal-processor/logsimulator"
)

func main() {
	dbType := "oracle"

	var parser dbparsers.LogParser
	switch dbType {
	case "oracle":
		parser = &dbparsers.OracleLogParser{}
	case "postgres":
		parser = &dbparsers.PostgresLogParser{}
	default:
		log.Fatal("Unsupported database type")
	}

	processor := logprocessor.SignalProcessor{}
	processor.AddGenerator(&logprocessor.FieldLevenshteinGenerator{FieldName: "bio"})

	logs := logsimulator.GenerateLogs(dbType, "UPDATE", "users", 5)

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
