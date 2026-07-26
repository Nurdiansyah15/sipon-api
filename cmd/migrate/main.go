package main

import (
	"fmt"
	"log"
	"os"
	"sipon-api/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate [up|down|fresh|version|force VERSION]")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gagal load config: %v", err)
	}

	m, err := migrate.New("file://"+cfg.Migration.MigrationsDir, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("gagal init migrator: %v", err)
	}
	defer m.Close()

	command := os.Args[1]
	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up gagal: %v", err)
		}
		log.Println("Migrasi UP berhasil")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate down gagal: %v", err)
		}
		log.Println("Migrasi DOWN berhasil")

	case "fresh":
		if err := m.Drop(); err != nil {
			log.Fatalf("migrate fresh gagal saat drop: %v", err)
		}
		srcErr, dbErr := m.Close()
		if srcErr != nil || dbErr != nil {
			log.Fatalf("migrate fresh gagal saat close migrator lama: sourceErr=%v dbErr=%v", srcErr, dbErr)
		}

		freshMigrator, err := migrate.New("file://"+cfg.Migration.MigrationsDir, cfg.Database.DSN)
		if err != nil {
			log.Fatalf("migrate fresh gagal init migrator baru: %v", err)
		}
		defer freshMigrator.Close()

		if err := freshMigrator.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate fresh gagal saat up: %v", err)
		}
		log.Println("Migrasi FRESH berhasil")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("gagal cek versi: %v", err)
		}
		fmt.Printf("Versi: %d (dirty: %v)\n", version, dirty)

	case "force":
		if len(os.Args) < 3 {
			fmt.Println("Usage: migrate force VERSION")
			os.Exit(1)
		}
		var version int
		fmt.Sscanf(os.Args[2], "%d", &version)
		if err := m.Force(version); err != nil {
			log.Fatalf("migrate force gagal: %v", err)
		}
		log.Printf("Force ke versi %d berhasil", version)

	default:
		fmt.Printf("Command tidak dikenal: %s\n", command)
		os.Exit(1)
	}
}
