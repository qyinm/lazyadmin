package db

import (
	"database/sql"
	"fmt"
	"net"

	"github.com/charmbracelet/bubbles/table"
	"github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qyinm/lazyadmin/config"
)

type Connection struct {
	DB     *sql.DB
	Tunnel *SSHTunnel
}

func Connect(cfg *config.DatabaseConfig) (*Connection, error) {
	var tunnel *SSHTunnel
	var dsn string
	var err error

	host := cfg.Host
	port := cfg.Port

	if cfg.SSH != nil {
		tunnel, err = NewSSHTunnel(cfg.SSH, cfg.Host, cfg.Port)
		if err != nil {
			return nil, fmt.Errorf("failed to create SSH tunnel: %w", err)
		}
		h, p, _ := net.SplitHostPort(tunnel.LocalAddr)
		host = h
		fmt.Sscanf(p, "%d", &port)
	}

	switch cfg.Driver {
	case "sqlite", "sqlite3", "":
		dsn = cfg.Path
		if dsn == "" {
			dsn = cfg.Name
		}
	case "postgres", "postgresql":
		sslMode := cfg.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			host, port, cfg.User, cfg.Password, cfg.Name, sslMode)
	case "mysql":
		mysqlCfg := mysql.NewConfig()
		mysqlCfg.Net = "tcp"
		mysqlCfg.Addr = fmt.Sprintf("%s:%d", host, port)
		mysqlCfg.User = cfg.User
		mysqlCfg.Passwd = cfg.Password
		mysqlCfg.DBName = cfg.Name
		dsn = mysqlCfg.FormatDSN()
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}

	driver := cfg.Driver
	if driver == "" || driver == "sqlite" {
		driver = "sqlite3"
	}
	if driver == "postgresql" {
		driver = "postgres"
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		if tunnel != nil {
			tunnel.Close()
		}
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		if tunnel != nil {
			tunnel.Close()
		}
		return nil, err
	}

	return &Connection{DB: db, Tunnel: tunnel}, nil
}

func (c *Connection) Close() error {
	if c.DB != nil {
		c.DB.Close()
	}
	if c.Tunnel != nil {
		c.Tunnel.Close()
	}
	return nil
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
