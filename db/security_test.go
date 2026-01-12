package db

import (
	"strings"
	"testing"
)

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		driver     string
		identifier string
		expected   string
	}{
		{
			name:       "postgres simple",
			driver:     "postgres",
			identifier: "users",
			expected:   `"users"`,
		},
		{
			name:       "postgres with double quote injection",
			driver:     "postgres",
			identifier: `users"; DROP TABLE users; --`,
			expected:   `"users""; DROP TABLE users; --"`,
		},
		{
			name:       "mysql simple",
			driver:     "mysql",
			identifier: "users",
			expected:   "`users`",
		},
		{
			name:       "mysql with backtick injection",
			driver:     "mysql",
			identifier: "users`; DROP TABLE users; --",
			expected:   "`users``; DROP TABLE users; --`",
		},
		{
			name:       "sqlite simple",
			driver:     "sqlite",
			identifier: "users",
			expected:   `"users"`,
		},
		{
			name:       "sqlite with injection attempt",
			driver:     "sqlite",
			identifier: `users"; DELETE FROM users; --`,
			expected:   `"users""; DELETE FROM users; --"`,
		},
		{
			name:       "empty driver defaults to double quotes",
			driver:     "",
			identifier: "table_name",
			expected:   `"table_name"`,
		},
		{
			name:       "reserved word",
			driver:     "postgres",
			identifier: "select",
			expected:   `"select"`,
		},
		{
			name:       "special characters",
			driver:     "postgres",
			identifier: "user-data",
			expected:   `"user-data"`,
		},
		{
			name:       "spaces in name",
			driver:     "mysql",
			identifier: "my table",
			expected:   "`my table`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := QuoteIdentifier(tt.driver, tt.identifier)
			if result != tt.expected {
				t.Errorf("QuoteIdentifier(%q, %q) = %q, want %q",
					tt.driver, tt.identifier, result, tt.expected)
			}
		})
	}
}

func TestEscapeSQLiteString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "users",
			expected: "users",
		},
		{
			name:     "single quote injection",
			input:    "users'; DROP TABLE users; --",
			expected: "users''; DROP TABLE users; --",
		},
		{
			name:     "multiple single quotes",
			input:    "it's a user's table",
			expected: "it''s a user''s table",
		},
		{
			name:     "no quotes",
			input:    "simple_table_name",
			expected: "simple_table_name",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only quote",
			input:    "'",
			expected: "''",
		},
		{
			name:     "consecutive quotes",
			input:    "'''",
			expected: "''''''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeSQLiteString(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeSQLiteString(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsSQLite(t *testing.T) {
	tests := []struct {
		driver   string
		expected bool
	}{
		{"sqlite", true},
		{"sqlite3", true},
		{"", true},
		{"postgres", false},
		{"postgresql", false},
		{"mysql", false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			result := isSQLite(tt.driver)
			if result != tt.expected {
				t.Errorf("isSQLite(%q) = %v, want %v",
					tt.driver, result, tt.expected)
			}
		})
	}
}

func TestBuildSelectAllQuery_Injection(t *testing.T) {
	tests := []struct {
		name      string
		driver    string
		tableName string
	}{
		{
			name:      "postgres injection attempt",
			driver:    "postgres",
			tableName: `users"; DROP TABLE users; --`,
		},
		{
			name:      "mysql injection attempt",
			driver:    "mysql",
			tableName: "users`; DROP TABLE users; --",
		},
		{
			name:      "sqlite injection attempt",
			driver:    "sqlite",
			tableName: `users"; DELETE FROM users; --`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := BuildSelectAllQuery(tt.driver, tt.tableName, 100)

			if strings.Contains(query, "DROP TABLE") && !strings.Contains(query, `"`) && !strings.Contains(query, "`") {
				t.Errorf("Query may be vulnerable to SQL injection: %s", query)
			}

			if strings.Count(query, "SELECT") != 1 {
				t.Errorf("Query should contain exactly one SELECT: %s", query)
			}
		})
	}
}

func TestQuoteIdentifier_NoUnquotedInjection(t *testing.T) {
	maliciousInputs := []string{
		`"; DROP TABLE users; --`,
		"`; DROP TABLE users; --",
		"'; DELETE FROM users; --",
		"table; INSERT INTO users VALUES(1); --",
		"users UNION SELECT * FROM passwords",
	}

	drivers := []string{"postgres", "mysql", "sqlite", ""}

	for _, driver := range drivers {
		for _, input := range maliciousInputs {
			t.Run(driver+"_"+input[:10], func(t *testing.T) {
				result := QuoteIdentifier(driver, input)

				if driver == "mysql" {
					if !strings.HasPrefix(result, "`") || !strings.HasSuffix(result, "`") {
						t.Errorf("MySQL identifier not properly quoted: %s", result)
					}
				} else {
					if !strings.HasPrefix(result, `"`) || !strings.HasSuffix(result, `"`) {
						t.Errorf("Identifier not properly quoted: %s", result)
					}
				}
			})
		}
	}
}
