package db

import (
	"database/sql"
	"fmt"
)

type TableInfo struct {
	Name   string
	Schema string
}

type ColumnInfo struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
	Default    sql.NullString
}

func GetTables(db *sql.DB, driver string) ([]TableInfo, error) {
	var query string

	switch {
	case isSQLite(driver):
		query = `SELECT name, '' as schema FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	case driver == "postgres" || driver == "postgresql":
		query = `SELECT table_name, table_schema FROM information_schema.tables 
				 WHERE table_schema NOT IN ('pg_catalog', 'information_schema') 
				 ORDER BY table_schema, table_name`
	case driver == "mysql":
		query = `SELECT table_name, table_schema FROM information_schema.tables 
				 WHERE table_schema = DATABASE() 
				 ORDER BY table_name`
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.Schema); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}

	return tables, rows.Err()
}

func GetColumns(db *sql.DB, driver, tableName string) ([]ColumnInfo, error) {
	return GetColumnsWithSchema(db, driver, tableName, "public")
}

func GetColumnsWithSchema(db *sql.DB, driver, tableName, schema string) ([]ColumnInfo, error) {
	var query string
	var args []interface{}

	switch {
	case isSQLite(driver):
		query = fmt.Sprintf(`PRAGMA table_info('%s')`, EscapeSQLiteString(tableName))
	case driver == "postgres" || driver == "postgresql":
		query = `SELECT 
					c.column_name,
					c.data_type,
					c.is_nullable = 'YES' as nullable,
					EXISTS (
						SELECT 1 
						FROM information_schema.table_constraints tc
						JOIN information_schema.key_column_usage kcu 
							ON tc.constraint_name = kcu.constraint_name
							AND tc.table_schema = kcu.table_schema
							AND tc.table_name = kcu.table_name
						WHERE tc.constraint_type = 'PRIMARY KEY'
							AND tc.table_schema = c.table_schema
							AND tc.table_name = c.table_name
							AND kcu.column_name = c.column_name
					) as is_pk,
					c.column_default
				FROM information_schema.columns c
				WHERE c.table_name = $1 AND c.table_schema = $2
				ORDER BY c.ordinal_position`
		args = []interface{}{tableName, schema}
	case driver == "mysql":
		query = `SELECT 
					COLUMN_NAME,
					DATA_TYPE,
					IS_NULLABLE = 'YES' as nullable,
					COLUMN_KEY = 'PRI' as is_pk,
					COLUMN_DEFAULT
				FROM information_schema.columns 
				WHERE table_schema = DATABASE() AND table_name = ?
				ORDER BY ordinal_position`
		args = []interface{}{tableName}
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo

	if isSQLite(driver) {
		for rows.Next() {
			var cid int
			var name, colType string
			var notNull, pk int
			var dflt sql.NullString
			if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
				return nil, err
			}
			columns = append(columns, ColumnInfo{
				Name:       name,
				Type:       colType,
				Nullable:   notNull == 0,
				PrimaryKey: pk == 1,
				Default:    dflt,
			})
		}
	} else {
		for rows.Next() {
			var c ColumnInfo
			if err := rows.Scan(&c.Name, &c.Type, &c.Nullable, &c.PrimaryKey, &c.Default); err != nil {
				return nil, err
			}
			columns = append(columns, c)
		}
	}

	return columns, rows.Err()
}

var ErrNoPrimaryKey = fmt.Errorf("no primary key found")

func GetPrimaryKey(db *sql.DB, driver, tableName string) (string, error) {
	columns, err := GetColumns(db, driver, tableName)
	if err != nil {
		return "", err
	}

	for _, col := range columns {
		if col.PrimaryKey {
			return col.Name, nil
		}
	}

	return "", fmt.Errorf("%w for table %q", ErrNoPrimaryKey, tableName)
}
