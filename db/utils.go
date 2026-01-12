package db

import "strings"

func isSQLite(driver string) bool {
	return driver == "sqlite" || driver == "sqlite3" || driver == ""
}

func QuoteIdentifier(driver, identifier string) string {
	switch driver {
	case "mysql":
		escaped := strings.ReplaceAll(identifier, "`", "``")
		return "`" + escaped + "`"
	default:
		escaped := strings.ReplaceAll(identifier, "\"", "\"\"")
		return "\"" + escaped + "\""
	}
}

func EscapeSQLiteString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
