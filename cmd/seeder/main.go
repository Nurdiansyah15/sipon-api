package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sipon-api/internal/config"
	"sipon-api/internal/seeders"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: seeder [all|NAME]")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gagal load config: %v", err)
	}

	db, err := sql.Open("pgx", cfg.Database.DSN)
	if err != nil {
		log.Fatalf("gagal koneksi database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("database tidak bisa dijangkau: %v", err)
	}

	command := os.Args[1]
	ctx := context.Background()

	switch command {
	case "all":
		if err := seeders.RunAll(ctx, db); err != nil {
			log.Fatalf("seeder all gagal: %v", err)
		}
		log.Println("seeder all berhasil")
	default:
		if err := seeders.RunByName(ctx, db, command); err != nil {
			log.Fatalf("seeder %s gagal: %v", command, err)
		}
		log.Printf("seeder %s berhasil", command)
	}
}
