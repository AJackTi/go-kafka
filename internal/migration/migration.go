package migration

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	defaultAttempts = 20
	defaultTimeout  = time.Second
)

// Run applies all pending migrations from sourcePath to databaseURL.
// databaseURL may be either a mysql-go driver DSN or a mysql:// URL.
func Run(databaseURL, sourcePath string) (runErr error) {
	if strings.TrimSpace(databaseURL) == "" {
		return errors.New("migration: database URL is required")
	}
	if strings.TrimSpace(sourcePath) == "" {
		return errors.New("migration: source path is required")
	}

	sourceURL := "file://" + filepath.ToSlash(strings.TrimSpace(sourcePath))
	if absolutePath, err := filepath.Abs(sourcePath); err == nil {
		sourceURL = "file://" + filepath.ToSlash(absolutePath)
	}

	var (
		migration *migrate.Migrate
		err       error
	)
	for attempts := defaultAttempts; attempts > 0; attempts-- {
		migration, err = migrate.New(sourceURL, migrationURL(databaseURL))
		if err == nil {
			break
		}
		if migration != nil {
			sourceErr, databaseErr := migration.Close()
			if sourceErr != nil || databaseErr != nil {
				err = errors.Join(err, sourceErr, databaseErr)
			}
		}
		if attempts > 1 {
			time.Sleep(defaultTimeout)
		}
	}
	if err != nil {
		return fmt.Errorf("migration: connect: %w", err)
	}
	if migration == nil {
		return errors.New("migration: constructor returned no migration")
	}
	defer func() {
		sourceErr, databaseErr := migration.Close()
		if sourceErr == nil && databaseErr == nil {
			return
		}
		closeErr := errors.Join(sourceErr, databaseErr)
		if runErr == nil {
			runErr = fmt.Errorf("migration: close: %w", closeErr)
			return
		}
		runErr = errors.Join(runErr, fmt.Errorf("migration: close: %w", closeErr))
	}()

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration: apply: %w", err)
	}

	return nil
}

func migrationURL(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "mysql://") {
		return value
	}
	return "mysql://" + value
}
