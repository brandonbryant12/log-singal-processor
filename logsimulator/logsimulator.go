package logsimulator

import (
	"fmt"
	"math/rand"
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

// GetDefaultFields returns the predefined field configurations.
// This allows the CLI to access the available fields.
func GetDefaultFields() []FieldConfig {
	return defaultFields
}

// GetFieldByName returns a field configuration by name
func GetFieldByName(name string) (FieldConfig, bool) {
	for _, field := range defaultFields {
		if field.Name == name {
			return field, true
		}
	}
	return FieldConfig{}, false
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
func GenerateLogs(dbType string, operation string, table string, numRows int, fields []FieldConfig, encConfig EncryptionConfig) []interface{} {
	logs := []interface{}{}

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

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
			beforeValue := field.Generator()
			afterValue := field.Generator()

			before[field.Name] = beforeValue

			// Potentially encrypt the after value based on configuration
			if encryptedValue, err := MaybeEncrypt(afterValue, encConfig); err == nil {
				after[field.Name] = encryptedValue
			} else {
				// If encryption fails, use the original value
				after[field.Name] = afterValue
			}
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
func GenerateDefaultLogs(dbType string, operation string, table string, numRows int, encConfig EncryptionConfig) []interface{} {
	return GenerateLogs(dbType, operation, table, numRows, defaultFields, encConfig)
}
