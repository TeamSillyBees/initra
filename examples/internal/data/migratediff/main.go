//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	atlas "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/lib/pq"
	"github.com/teamsillybees/initra/examples/internal/boot"
	"github.com/teamsillybees/initra/examples/internal/data/ent/migrate"
)

const defaultConfigDir = "configs"

type migrateOptions struct {
	name      string
	devURL    string
	env       string
	configDir string
}

func main() {
	options, err := parseArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("usage: go run ./internal/data/migratediff/main.go <name> [-env <env>] [-config-dir <dir>] [-dev-url <url>]: %v", err)
	}
	devURL, err := resolveDevURL(options)
	if err != nil {
		log.Fatalf("failed resolving migration database: %v", err)
	}
	dir, err := atlas.NewLocalDir("db/migrations")
	if err != nil {
		log.Fatalf("failed opening migration directory: %v", err)
	}
	opts := []schema.MigrateOption{
		schema.WithDir(dir),
		schema.WithMigrationMode(schema.ModeReplay),
		schema.WithDialect(dialect.Postgres),
		schema.WithFormatter(atlas.DefaultFormatter),
		migrate.WithForeignKeys(false),
	}
	if err := migrate.NamedDiff(context.Background(), devURL, options.name, opts...); err != nil {
		log.Fatalf("failed generating migration file: %v", err)
	}
}

func parseArgs(args []string) (migrateOptions, error) {
	options := migrateOptions{configDir: defaultConfigDir}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-dev-url" || arg == "--dev-url":
			i++
			if i >= len(args) || strings.TrimSpace(args[i]) == "" {
				return migrateOptions{}, fmt.Errorf("%s requires a value", arg)
			}
			options.devURL = args[i]
		case strings.HasPrefix(arg, "-dev-url="):
			value := strings.TrimPrefix(arg, "-dev-url=")
			if strings.TrimSpace(value) == "" {
				return migrateOptions{}, fmt.Errorf("-dev-url requires a value")
			}
			options.devURL = value
		case strings.HasPrefix(arg, "--dev-url="):
			value := strings.TrimPrefix(arg, "--dev-url=")
			if strings.TrimSpace(value) == "" {
				return migrateOptions{}, fmt.Errorf("--dev-url requires a value")
			}
			options.devURL = value
		case arg == "-env" || arg == "--env":
			i++
			if i >= len(args) || strings.TrimSpace(args[i]) == "" {
				return migrateOptions{}, fmt.Errorf("%s requires a value", arg)
			}
			options.env = args[i]
		case arg == "-e":
			i++
			if i >= len(args) || strings.TrimSpace(args[i]) == "" {
				return migrateOptions{}, fmt.Errorf("%s requires a value", arg)
			}
			options.env = args[i]
		case strings.HasPrefix(arg, "-env="):
			value := strings.TrimPrefix(arg, "-env=")
			if strings.TrimSpace(value) == "" {
				return migrateOptions{}, fmt.Errorf("-env requires a value")
			}
			options.env = value
		case strings.HasPrefix(arg, "--env="):
			value := strings.TrimPrefix(arg, "--env=")
			if strings.TrimSpace(value) == "" {
				return migrateOptions{}, fmt.Errorf("--env requires a value")
			}
			options.env = value
		case arg == "-config-dir" || arg == "--config-dir":
			i++
			if i >= len(args) || strings.TrimSpace(args[i]) == "" {
				return migrateOptions{}, fmt.Errorf("%s requires a value", arg)
			}
			options.configDir = args[i]
		case strings.HasPrefix(arg, "-config-dir="):
			value := strings.TrimPrefix(arg, "-config-dir=")
			if strings.TrimSpace(value) == "" {
				return migrateOptions{}, fmt.Errorf("-config-dir requires a value")
			}
			options.configDir = value
		case strings.HasPrefix(arg, "--config-dir="):
			value := strings.TrimPrefix(arg, "--config-dir=")
			if strings.TrimSpace(value) == "" {
				return migrateOptions{}, fmt.Errorf("--config-dir requires a value")
			}
			options.configDir = value
		case strings.HasPrefix(arg, "-"):
			return migrateOptions{}, fmt.Errorf("unknown option %q", arg)
		case options.name == "":
			options.name = arg
		default:
			return migrateOptions{}, fmt.Errorf("only one migration name is allowed")
		}
	}
	if strings.TrimSpace(options.name) == "" {
		return migrateOptions{}, fmt.Errorf("migration name is required")
	}
	return options, nil
}

func resolveDevURL(options migrateOptions) (string, error) {
	if strings.TrimSpace(options.devURL) != "" {
		return options.devURL, nil
	}
	cfg, err := boot.LoadConfig(options.env, options.configDir)
	if err != nil {
		return "", err
	}
	return databaseURL(cfg), nil
}

func databaseURL(cfg *boot.Config) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   databaseUser(cfg.Database.User, cfg.Database.Password),
		Host:   net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port)),
		Path:   "/" + cfg.Database.DBName,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return u.String()
}

func databaseUser(user string, password string) *url.Userinfo {
	if password == "" {
		return url.User(user)
	}
	return url.UserPassword(user, password)
}
