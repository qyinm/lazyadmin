package main

import (
	"fmt"
	"log"

	"github.com/qyinm/lazyadmin/config"
	"github.com/qyinm/lazyadmin/db"
)

func main() {
	cfg, err := config.Load("test/admin-ssh.yaml")
	if err != nil {
		log.Fatal("Config error:", err)
	}

	fmt.Println("Connecting via SSH tunnel...")
	conn, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer conn.Close()

	fmt.Println("Connected! Running query...")
	cols, rows, err := db.RunQuery(conn.DB, "SELECT id, email, role FROM users")
	if err != nil {
		log.Fatal("Query error:", err)
	}

	fmt.Println("\nColumns:", len(cols))
	for _, c := range cols {
		fmt.Printf("  - %s\n", c.Title)
	}

	fmt.Println("\nRows:", len(rows))
	for _, r := range rows {
		fmt.Printf("  %v\n", r)
	}

	fmt.Println("\n✅ SSH tunnel test passed!")
}
