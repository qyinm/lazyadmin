package db

import (
	"database/sql"
	"fmt"
	"strings"
)

func SelectAll(db *sql.DB, tableName string, limit int) (string, error) {
	if limit <= 0 {
		limit = 100
	}
	return fmt.Sprintf("SELECT * FROM %s LIMIT %d", tableName, limit), nil
}

func InsertRecord(db *sql.DB, driver, tableName string, data map[string]interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("no data to insert")
	}

	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	i := 1
	for col, val := range data {
		columns = append(columns, col)
		values = append(values, val)

		switch driver {
		case "postgres", "postgresql":
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		default:
			placeholders = append(placeholders, "?")
		}
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	_, err := db.Exec(query, values...)
	return err
}

func UpdateRecord(db *sql.DB, driver, tableName, pkColumn string, pkValue interface{}, data map[string]interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("no data to update")
	}

	setClauses := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+1)

	i := 1
	for col, val := range data {
		var placeholder string
		switch driver {
		case "postgres", "postgresql":
			placeholder = fmt.Sprintf("$%d", i)
		default:
			placeholder = "?"
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", col, placeholder))
		values = append(values, val)
		i++
	}

	var pkPlaceholder string
	switch driver {
	case "postgres", "postgresql":
		pkPlaceholder = fmt.Sprintf("$%d", i)
	default:
		pkPlaceholder = "?"
	}
	values = append(values, pkValue)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = %s",
		tableName,
		strings.Join(setClauses, ", "),
		pkColumn,
		pkPlaceholder)

	result, err := db.Exec(query, values...)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("no rows updated")
	}

	return nil
}

func DeleteRecord(db *sql.DB, driver, tableName, pkColumn string, pkValue interface{}) error {
	var placeholder string
	switch driver {
	case "postgres", "postgresql":
		placeholder = "$1"
	default:
		placeholder = "?"
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = %s", tableName, pkColumn, placeholder)

	result, err := db.Exec(query, pkValue)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("no rows deleted")
	}

	return nil
}

func GetRecordByPK(db *sql.DB, driver, tableName, pkColumn string, pkValue interface{}) (map[string]interface{}, error) {
	var placeholder string
	switch driver {
	case "postgres", "postgresql":
		placeholder = "$1"
	default:
		placeholder = "?"
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = %s", tableName, pkColumn, placeholder)

	rows, err := db.Query(query, pkValue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, fmt.Errorf("record not found")
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	record := make(map[string]interface{})
	for i, col := range columns {
		record[col] = values[i]
	}

	return record, nil
}
