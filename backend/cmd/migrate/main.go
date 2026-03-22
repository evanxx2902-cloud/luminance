package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		command    = flag.String("command", "up", "Migration command: up, down, force, version")
		versionArg = flag.String("version", "", "Version for force command")
		dbType     = flag.String("db", "business", "Database type: business or vector")
	)
	flag.Parse()

	// 从环境变量读取连接字符串，或使用默认值
	var dbURL string
	var migrationsPath string

	switch *dbType {
	case "business":
		dbURL = getEnv("DATABASE_BUSINESS_URL", "postgres://luminance:luminance@127.0.0.1:5432/luminance?sslmode=disable")
		migrationsPath = "file://backend/migrations"
	case "vector":
		dbURL = getEnv("DATABASE_VECTOR_URL", "postgres://luminance:luminance@127.0.0.1:5433/luminance_vector?sslmode=disable")
		// 向量库只执行 000002 迁移
		migrationsPath = "file://backend/migrations"
	default:
		fmt.Fprintf(os.Stderr, "Unknown db type: %s\n", *dbType)
		os.Exit(1)
	}

	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create migrate instance: %v\n", err)
		os.Exit(1)
	}

	switch *command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			fmt.Fprintf(os.Stderr, "Migration up failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migration up completed")

	case "down":
		if err := m.Steps(-1); err != nil {
			fmt.Fprintf(os.Stderr, "Migration down failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migration down completed")

	case "force":
		if *versionArg == "" {
			fmt.Fprintln(os.Stderr, "Version required for force command")
			os.Exit(1)
		}
		v, err := strconv.Atoi(*versionArg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid version: %v\n", err)
			os.Exit(1)
		}
		if err := m.Force(v); err != nil {
			fmt.Fprintf(os.Stderr, "Force migration failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Force migration to version %d completed\n", v)

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", v, dirty)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", *command)
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
