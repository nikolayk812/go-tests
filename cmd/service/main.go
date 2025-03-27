package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nikolayk812/go-tests/internal/repository"
	"github.com/nikolayk812/go-tests/internal/rest"
	"github.com/nikolayk812/go-tests/internal/service"
	"github.com/shopspring/decimal"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/*
docker run -d -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password -e POSTGRES_DB=dbname -p 5432:5432 postgres:17.2-alpine
*/
func main() {
	//gin.SetMode(gin.ReleaseMode)
	decimal.MarshalJSONWithoutQuotes = true

	connStr := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	var gErr error
	defer func() {
		if gErr != nil {
			slog.Error("startup error", "err", gErr)
			os.Exit(1)
		}

		os.Exit(0)
	}()

	if err := runMigrations(connStr); err != nil {
		gErr = fmt.Errorf("runMigrations: %w", err)
		return
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		gErr = fmt.Errorf("pgxpool.New: %w", err)
		return
	}

	repo, err := repository.New(pool)
	if err != nil {
		gErr = fmt.Errorf("repository.New: %w", err)
		return
	}

	cartService, err := service.NewCart(repo)
	if err != nil {
		gErr = fmt.Errorf("service.NewCart: %w", err)
		return
	}

	cartHandler, err := rest.NewCart(cartService)
	if err != nil {
		gErr = fmt.Errorf("rest.NewCart: %w", err)
		return
	}

	router := rest.SetupRouter(cartHandler)

	if err := runServer(ctx, router); err != nil {
		gErr = fmt.Errorf("runServer: %w", err)
		return
	}
}

func runServer(ctx context.Context, handler http.Handler) error {
	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Channel to signal server start errors
	serverErr := make(chan error, 1)

	// Run server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-stop:
		// Create a context with a timeout for the graceful shutdown
		ctxShutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctxShutdown); err != nil {
			return fmt.Errorf("server.Shutdown: %w", err)
		}

		slog.Info("server gracefully stopped")
		return nil
	case err := <-serverErr:
		return fmt.Errorf("server.ListenAndServe: %w", err)
	}
}

func runMigrations(connStr string) error {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres.WithInstance: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./internal/repository/migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate.NewWithDatabaseInstance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("m.Up: %w", err)
	}

	return nil
}
