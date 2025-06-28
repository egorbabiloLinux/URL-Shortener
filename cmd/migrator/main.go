package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		migrationsPath 	string 
		migrationsTable string 
		doDown 			bool
		force 			int
	)

	flag.StringVar(&migrationsPath, "migrations-path", "", "Path to the migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "Name of migrations table")
	flag.BoolVar(&doDown, "down", false, "Run down migrations")
	flag.IntVar(&force, "force", 0, "Set migrations version")
	flag.Parse()

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		panic("DATABASE_URL is required")
	}

	dsn := fmt.Sprintf("%s&x-migrations-table=%s", url, migrationsTable)

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		panic(err)
	}

	if force > 0 {
		if err := m.Force(force); err != nil {
			panic(err)
		}
	}

	if doDown {
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("No migrations to apply")

				return
			}

			panic(err)
		}
	} else {
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("No migrations to apply")

				return 
			}

			panic(err)
		}
	}
}