//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/qyinm/lazyadmin/config"
	"github.com/qyinm/lazyadmin/db"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run test_connection.go <config.yaml>")
	}

	configPath := os.Args[1]

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal("Config error:", err)
	}

	fmt.Printf("Connecting to %s database...\n", cfg.Database.Driver)

	conn, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer conn.Close()

	fmt.Println("Connected successfully!")

	if len(cfg.Views) > 0 {
		fmt.Printf("Running test query: %s\n", cfg.Views[0].Title)
		cols, rows, err := db.RunQuery(conn.DB, cfg.Views[0].Query)
		if err != nil {
			log.Fatal("Query error:", err)
		}

		fmt.Printf("Columns: %d, Rows: %d\n", len(cols), len(rows))

		for _, c := range cols {
			fmt.Printf("  - %s\n", c.Title)
		}
	}

	fmt.Println("✅ Connection test passed!")
}
