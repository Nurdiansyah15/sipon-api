package testhelper

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"sipon-api/internal/migrations"
)

// MustStartTestDB starts a postgres:16-alpine container, runs all migrations,
// and returns a ready *sql.DB. Call once from TestMain; defer the cleanup func.
func MustStartTestDB() (*sql.DB, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image: "postgres:16-alpine",
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("testhelper: start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("testhelper: get container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("testhelper: get mapped port: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=test password=test dbname=testdb sslmode=disable",
		host, port.Port(),
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("testhelper: open db: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for {
		if err := db.PingContext(pingCtx); err == nil {
			break
		}
		select {
		case <-pingCtx.Done():
			log.Fatalf("testhelper: db never became ready: %v", pingCtx.Err())
		case <-time.After(200 * time.Millisecond):
		}
	}

	migrateDSN := fmt.Sprintf(
		"postgres://test:test@%s:%s/testdb?sslmode=disable",
		host, port.Port(),
	)
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		log.Fatalf("testhelper: create migration source: %v", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, migrateDSN)
	if err != nil {
		log.Fatalf("testhelper: create migrator: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("testhelper: run migrations: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
		if err := container.Terminate(ctx); err != nil {
			log.Printf("testhelper: terminate container: %v", err)
		}
	}
	return db, cleanup
}
