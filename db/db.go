package db

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	_ "github.com/mattn/go-sqlite3"
)

func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func RunQuery(db *sql.DB, query string) ([]table.Column, []table.Row, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	columns := make([]table.Column, len(columnNames))
	for i, name := range columnNames {
		width := len(name)
		if width < 10 {
			width = 10
		}
		columns[i] = table.Column{Title: name, Width: width}
	}

	var tableRows []table.Row
	for rows.Next() {
		values := make([]interface{}, len(columnNames))
		valuePtrs := make([]interface{}, len(columnNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, err
		}

		row := make(table.Row, len(columnNames))
		for i, val := range values {
			row[i] = toString(val)
		}
		tableRows = append(tableRows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return columns, tableRows, nil
}

func toString(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	switch v := val.(type) {
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
