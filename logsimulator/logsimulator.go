package logsimulator

import (
	"fmt"
	"time"
)

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

func GenerateLogs(dbType string, operation string, table string, numRows int) []interface{} {
	logs := []interface{}{}
	for i := 1; i <= numRows; i++ {
		rowID := fmt.Sprintf("row%d", i)
		before := map[string]interface{}{"bio": "original bio", "email": "original@example.com"}
		after := map[string]interface{}{"bio": "encrypted_data", "email": "hidden"}
		var log interface{}
		if dbType == "oracle" {
			log = GenerateOracleUpdateLog(table, rowID, []string{"bio", "email"}, before, after)
		} else if dbType == "postgres" {
			log = GeneratePostgresUpdateLog(table, rowID, []string{"bio", "email"}, before, after)
		}
		logs = append(logs, log)
	}
	return logs
}
