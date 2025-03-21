
package logsimulator

import (
	"fmt"
	"time"
	"github.com/brianvoe/gofakeit/v7"
)

// FieldConfig defines a field name and its corresponding data generator function.
// The Generator function returns a string, aligning with most gofakeit functions.
type FieldConfig struct {
	Name      string
	Generator func() string
}

// defaultFields provides a set of predefined fields with generators for common use cases.
var defaultFields = []FieldConfig{
	{Name: "bio", Generator: func() string { return gofakeit.Sentence(5) }},
	{Name: "email", Generator: gofakeit.Email},
	{Name: "phone", Generator: gofakeit.Phone},
	{Name: "address", Generator: func() string { return gofakeit.Address().Address }},
}

// GenerateOracleUpdateLog creates a mock log entry for an Oracle UPDATE operation.
func GenerateOracleUpdateLog(table string, rowID string, columns []string, before map[string]interface{}, after map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"action":          "UPDATE",
		"table_name":      table,
		"rowid":           rowID,
		"changed_columns": columns,
		"timestamp":       time.Now(),
		"before_values":   before,
		"after_values":    after,
	}
}

// GeneratePostgresUpdateLog creates a mock log entry for a PostgreSQL UPDATE operation.
func GeneratePostgresUpdateLog(table string, primaryKey string, columns []string, before map[string]interface{}, after map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"operation":       "UPDATE",
		"table":           table,
		"primary_key":     primaryKey,
		"changed_columns": columns,
		"timestamp":       time.Now(),
		"old_values":      before,
		"new_values":      after,
	}
}

// GenerateLogs generates a specified number of mock log entries based on the database type,
// operation, table, and field configurations.
func GenerateLogs(dbType string, operation string, table string, numRows int, fields []FieldConfig) []interface{} {
	logs := []interface{}{}

	// Extract field names to use as columns
	columns := make([]string, len(fields))
	for i, field := range fields {
		columns[i] = field.Name
	}

	// Generate the specified number of log entries
	for i := 1; i <= numRows; i++ {
		rowID := fmt.Sprintf("row%d", i)
		before := make(map[string]interface{})
		after := make(map[string]interface{})

		// Populate before and after values using the field generators
		for _, field := range fields {
			before[field.Name] = field.Generator()
			after[field.Name] = field.Generator()
		}

		// Generate the log based on the database type
		var log interface{}
		if dbType == "oracle" {
			log = GenerateOracleUpdateLog(table, rowID, columns, before, after)
		} else if dbType == "postgres" {
			log = GeneratePostgresUpdateLog(table, rowID, columns, before, after)
		}
		logs = append(logs, log)
	}
	return logs
}

// GenerateDefaultLogs generates a specified number of mock log entries using the default field configurations.
func GenerateDefaultLogs(dbType string, operation string, table string, numRows int) []interface{} {
	return GenerateLogs(dbType, operation, table, numRows, defaultFields)
}
      