package migrate

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Entry struct {
	Version string
	SQL     string
}

var migrations = []Entry{
	{
		Version: "v0.0.1",
		SQL: `
			CREATE TABLE IF NOT EXISTS "rankguessr_migrations" (
				"version" TEXT PRIMARY KEY
			);

			CREATE TABLE IF NOT EXISTS "players" (
				"osu_id" INTEGER PRIMARY KEY,
				"source" TEXT NOT NULL,
				"created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				"checked_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);
			
			CREATE TABLE IF NOT EXISTS "users" (
				"osu_id" INTEGER PRIMARY KEY,
				"username" TEXT NOT NULL,
				"avatar_url" TEXT NOT NULL,
				"country_code" TEXT NOT NULL,
				"created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				"updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);

			CREATE TABLE IF NOT EXISTS "guesses" (
				"id" CHAR(27) PRIMARY KEY,
				"user_id" INTEGER NOT NULL,
				"player_id" INTEGER NOT NULL,
				"elo" INTEGER NOT NULL,
				"guess" INTEGER NOT NULL,
				"actual_rank" INTEGER NOT NULL,
				"created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				"updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				FOREIGN KEY (player_id) REFERENCES players(osu_id) ON DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(osu_id) ON DELETE CASCADE
			);

			CREATE TABLE IF NOT EXISTS "rooms" (
				"id" CHAR(27) PRIMARY KEY,
				"user_id" INTEGER NOT NULL,
				"score_id" BIGINT NOT NULL,
				"player_id" INTEGER NOT NULL,
				"is_closed" BOOLEAN NOT NULL DEFAULT FALSE,
				"created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				"updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				FOREIGN KEY (user_id) REFERENCES users(osu_id) ON DELETE CASCADE,
				FOREIGN KEY (player_id) REFERENCES players(osu_id) ON DELETE CASCADE
			);

			CREATE TABLE IF NOT EXISTS "sessions" (
				"id" CHAR(27) PRIMARY KEY,
				"user_id" INTEGER NOT NULL UNIQUE,
				"access_token" TEXT NOT NULL,
				"refresh_token" TEXT NOT NULL,
				"expires_at" TIMESTAMPTZ NOT NULL,
				"created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				FOREIGN KEY (user_id) REFERENCES users(osu_id) ON DELETE CASCADE
			);
		`,
	},
	{
		Version: "v0.0.2",
		SQL: `
			ALTER TABLE "rooms" DROP COLUMN "is_closed";
			ALTER TABLE "rooms" ADD COLUMN "guess_id" CHAR(27);
			ALTER TABLE "rooms" ADD FOREIGN KEY (guess_id) REFERENCES guesses(id) ON DELETE CASCADE;
		`,
	},
	{
		Version: "v0.0.3",
		SQL: `
			ALTER TABLE "sessions" ADD UNIQUE (user_id);
		`,
	},
	{
		Version: "v0.0.7",
		SQL: `
			ALTER TABLE "users" ADD COLUMN "elo" INTEGER NOT NULL DEFAULT 0;
			ALTER TABLE "guesses" ADD COLUMN "beatmapset_id" INTEGER NOT NULL;
			ALTER TABLE "guesses" ADD COLUMN "beatmap_id" INTEGER NOT NULL;
			ALTER TABLE "guesses" ADD COLUMN "score_id" BIGINT NOT NULL;
		`,
	},
	{
		Version: "v0.0.8",
		SQL: `
			ALTER TABLE "sessions" ALTER COLUMN access_token TYPE BYTEA USING access_token::bytea;
			ALTER TABLE "sessions" ALTER COLUMN refresh_token TYPE BYTEA USING refresh_token::bytea;
		`,
	},
	{
		Version: "v0.0.9",
		SQL: `
			ALTER TABLE "guesses" DROP CONSTRAINT guesses_player_id_fkey;
		`,
	},
	{
		Version: "v0.1.0",
		SQL: `
			ALTER TABLE "rooms" DROP CONSTRAINT rooms_player_id_fkey;
		`,
	},
}

var latest = migrations[len(migrations)-1].Version

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	version := ""
	fromIdx := 0

	err := pool.QueryRow(ctx, "SELECT version FROM rankguessr_migrations").Scan(&version)
	if err == nil && version == latest {
		return nil
	}

	// only run if migration exists, else run all migrations from idx 0
	if version != "" {
		// find the current migration
		idx := slices.IndexFunc(migrations, func(m Entry) bool {
			return m.Version == version
		})
		if idx == -1 {
			return errors.New("invalid migration version")
		}

		fromIdx = idx + 1
	}

	return runMigrationsFromIdx(ctx, pool, fromIdx)
}

func runMigrationsFromIdx(ctx context.Context, pool *pgxpool.Pool, idx int) error {
	for _, m := range migrations[idx:] {
		_, err := pool.Exec(ctx, m.SQL)
		if err != nil {
			return fmt.Errorf("migration %s failed: %s", m.Version, err.Error())
		} else {
			fmt.Printf("migration %s successful\n", m.Version)
		}
	}

	_, err := pool.Exec(ctx, "DELETE FROM rankguessr_migrations")
	if err != nil {
		return fmt.Errorf("failed to clear migrations table: %s", err.Error())
	}

	_, err = pool.Exec(ctx, "INSERT INTO rankguessr_migrations (version) VALUES ($1)", latest)

	return err
}
