package db

import (
	"fmt"
	"strings"
)

func isSQLite(driver string) bool {
	return driver == "sqlite" || driver == "sqlite3"
}

func ValidateDriver(driver string) error {
	validDrivers := []string{"sqlite", "sqlite3", "postgres", "postgresql", "mysql"}
	for _, v := range validDrivers {
		if driver == v {
			return nil
		}
	}
	if driver == "" {
		return fmt.Errorf("database driver is required; supported drivers: sqlite, sqlite3, postgres, postgresql, mysql")
	}
	return fmt.Errorf("unsupported database driver %q; supported drivers: sqlite, sqlite3, postgres, postgresql, mysql", driver)
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
