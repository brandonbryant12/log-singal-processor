package dbparsers

import (
	"errors"
	"log-signal-processor/logprocessor"
	"time"
)

type LogParser interface {
	ParseLog(rawLog interface{}) (logprocessor.LogData, error)
}

type OracleLogParser struct{}

func (p *OracleLogParser) ParseLog(rawLog interface{}) (logprocessor.LogData, error) {
	logMap, ok := rawLog.(map[string]interface{})
	if !ok {
		return logprocessor.LogData{}, errors.New("invalid log format")
	}
	operation, _ := logMap["action"].(string)
	table, _ := logMap["table_name"].(string)
	rowID, _ := logMap["rowid"].(string)
	columns, _ := logMap["changed_columns"].([]string)
	timestamp, _ := logMap["timestamp"].(time.Time)
	before, _ := logMap["before_values"].(map[string]interface{})
	after, _ := logMap["after_values"].(map[string]interface{})
	return logprocessor.LogData{
		Operation:     operation,
		Table:         table,
		RowIdentifier: rowID,
		Columns:       columns,
		Timestamp:     timestamp,
		Before:        before,
		After:         after,
	}, nil
}

type PostgresLogParser struct{}

func (p *PostgresLogParser) ParseLog(rawLog interface{}) (logprocessor.LogData, error) {
	logMap, ok := rawLog.(map[string]interface{})
	if !ok {
		return logprocessor.LogData{}, errors.New("invalid log format")
	}
	operation, _ := logMap["operation"].(string)
	table, _ := logMap["table"].(string)
	primaryKey, _ := logMap["primary_key"].(string)
	columns, _ := logMap["changed_columns"].([]string)
	timestamp, _ := logMap["timestamp"].(time.Time)
	before, _ := logMap["old_values"].(map[string]interface{})
	after, _ := logMap["new_values"].(map[string]interface{})
	return logprocessor.LogData{
		Operation:     operation,
		Table:         table,
		RowIdentifier: primaryKey,
		Columns:       columns,
		Timestamp:     timestamp,
		Before:        before,
		After:         after,
	}, nil
}
