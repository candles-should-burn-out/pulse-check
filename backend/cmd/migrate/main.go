package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const defaultMigrationsDir = "migrations"

func main() {
	var migrationsDir string
	var timeout time.Duration
	var verbose bool

	flag.StringVar(&migrationsDir, "dir", envString("MIGRATIONS_DIR", ""), "directory with goose SQL migrations")
	flag.DurationVar(&timeout, "timeout", 30*time.Second, "database operation timeout")
	flag.BoolVar(&verbose, "v", false, "enable goose verbose logging")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(2)
	}

	if migrationsDir == "" {
		var err error
		migrationsDir, err = resolveMigrationsDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "resolve migrations dir: %v\n", err)
			os.Exit(2)
		}
	}

	if err := goose.SetDialect("postgres"); err != nil {
		fmt.Fprintf(os.Stderr, "set goose dialect: %v\n", err)
		os.Exit(1)
	}
	goose.SetVerbose(verbose)

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ping database: %v\n", err)
		os.Exit(1)
	}

	command := args[0]
	commandArgs := args[1:]
	if err := goose.RunContext(ctx, command, db, migrationsDir, commandArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "run migration command %q: %v\n", command, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
  pulse-check-migrate [flags] <command> [args]

Commands are passed through to goose, for example:
  up
  status
  version
  down
  down-to <version>

Environment:
  DATABASE_URL     PostgreSQL connection string.
  MIGRATIONS_DIR  Optional migrations directory.

`)
	flag.PrintDefaults()
}

func resolveMigrationsDir() (string, error) {
	if isDir(defaultMigrationsDir) {
		return defaultMigrationsDir, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(filepath.Dir(executable), defaultMigrationsDir)
	if isDir(dir) {
		return dir, nil
	}

	return "", errors.New("migrations directory not found; set MIGRATIONS_DIR or run from backend directory")
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func envString(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	return value
}
