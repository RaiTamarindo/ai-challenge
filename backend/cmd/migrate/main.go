package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

func main() {
	var direction string
	flag.StringVar(&direction, "direction", "up", "Migration direction: up or down")
	flag.Parse()

	// Get database connection string from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Open database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Configure migrations
	migrations := &migrate.FileMigrationSource{
		Dir: "/migrations",
	}

	var n int
	switch direction {
	case "up":
		n, err = migrate.Exec(db, "postgres", migrations, migrate.Up)
		if err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		fmt.Printf("Applied %d migrations\n", n)
	case "down":
		n, err = migrate.ExecMax(db, "postgres", migrations, migrate.Down, 1)
		if err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		fmt.Printf("Rolled back %d migration(s)\n", n)
	case "status":
		records, err := migrate.GetMigrationRecords(db, "postgres")
		if err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
		
		fmt.Println("Migration Status:")
		fmt.Println("================")
		
		if len(records) == 0 {
			fmt.Println("No migrations applied")
			return
		}
		
		for _, record := range records {
			fmt.Printf("Applied: %s at %v\n", record.Id, record.AppliedAt)
		}
	default:
		log.Fatalf("Invalid direction: %s. Use 'up', 'down', or 'status'", direction)
	}
}