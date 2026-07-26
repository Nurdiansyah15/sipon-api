package seeders

import (
	"context"
	"database/sql"
	"fmt"
)

type Seeder interface {
	Name() string
	Run(ctx context.Context, db *sql.DB) error
}

var registeredSeeders = []Seeder{
	RoleSeeder{},
	UserSeeder{},
}

func RunAll(ctx context.Context, db *sql.DB) error {
	for _, seeder := range registeredSeeders {
		if err := seeder.Run(ctx, db); err != nil {
			return fmt.Errorf("run seeder %s: %w", seeder.Name(), err)
		}
	}
	return nil
}

func RunByName(ctx context.Context, db *sql.DB, name string) error {
	for _, s := range registeredSeeders {
		if s.Name() == name {
			if err := s.Run(ctx, db); err != nil {
				return fmt.Errorf("run seeder %s: %w", s.Name(), err)
			}
			return nil
		}
	}
	return fmt.Errorf("seeder %q tidak ditemukan", name)
}
